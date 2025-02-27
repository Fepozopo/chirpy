package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"

	"github.com/Fepozopo/chirpy/internal/auth"
	"github.com/Fepozopo/chirpy/internal/database"
)

// HandleHealthz is a simple health-check endpoint that responds with a 200 OK and the
// plain text string "OK" to indicate that the server is running.
func (cfg *ApiConfig) HandleHealthz(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK\n"))
}

// HandleMetrics responds with a simple HTML page displaying the current value of the
// file server hit counter.
func (cfg *ApiConfig) HandleMetrics(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	hits := cfg.fileserverHits.Load()
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "<html>\n<body>\n<h1>Welcome, Chirpy Admin</h1>\n<p>Chirpy has been visited %d times!</p>\n</body>\n</html>\n", hits)
}

// HandleReset is a special admin-only endpoint that can only be accessed in a local
// development environment. It resets the server's hits counter to 0 and deletes all
// users in the database. It responds with a 200 OK and a plaintext message
// indicating that the hits counter has been reset to 0.
func (cfg *ApiConfig) HandleReset(w http.ResponseWriter, r *http.Request) {
	godotenv.Load()
	platform := os.Getenv("PLATFORM")
	apiCfg := ApiConfig{
		Platform: platform,
	}
	// Ensure this endpoint can only be accessed in a local development environment
	if apiCfg.Platform != "dev" {
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte("403 Forbidden\n"))
		return
	}

	// Delete all users in the database
	err := cfg.DbQueries.DeleteAllUsers(r.Context())
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Failed delete users"})
		return
	}

	// Reset the server hits count to 0
	cfg.fileserverHits.Store(0)

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Hits reset to 0 and all users deleted\n"))
}

// HandleCreateChirp processes a request to create a new chirp. It parses the request
// body into a CreateChirpRequest struct, validates the request with a JWT extracted
// from the Authorization header, checks the chirp for a maximum length and removes
// any profane words, and then stores the chirp in the database. If successful, it
// returns a 201 status code with the chirp data; otherwise, it returns an error
// status code with an appropriate error message.
func (cfg *ApiConfig) HandleCreateChirp(w http.ResponseWriter, r *http.Request) {
	// Parse the JSON body of the request into a CreateChirpRequest struct
	var createChirpRequest CreateChirpRequest
	if err := json.NewDecoder(r.Body).Decode(&createChirpRequest); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Invalid request body"})
		return
	}

	// Get the Bearer token from the request headers
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Invalid authorization header"})
		return
	}

	// Validate the JWT
	userID, err := auth.ValidateJWT(token, cfg.TokenSecret)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Invalid JWT"})
		return
	}

	// Set the user ID in the request body
	createChirpRequest.UserID = userID

	// Check if the chirp exceeds the 140 character limit
	if len(createChirpRequest.Body) > 140 {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Chirp is too long"})
		return
	}

	// Define the list of profane words to replace
	profaneWords := []string{"kerfuffle", "sharbert", "fornax"}

	// Replace profane words with **** in a case-insensitive manner
	cleanedBody := createChirpRequest.Body
	caser := cases.Title(language.English)
	for _, word := range profaneWords {
		cleanedBody = strings.ReplaceAll(cleanedBody, word, "****")
		cleanedBody = strings.ReplaceAll(cleanedBody, caser.String(word), "****")
		cleanedBody = strings.ReplaceAll(cleanedBody, strings.ToUpper(word), "****")
	}
	createChirpRequest.Body = cleanedBody

	// If the chirp is valid, save it in the database
	chirp, err := cfg.DbQueries.CreateChirp(r.Context(), database.CreateChirpParams(createChirpRequest))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Failed to create chirp"})
		return
	}

	// Map the chirp struct to a MappedChirp struct to control the JSON keys
	mappedChirp := MappedChirp{
		ID:        chirp.ID,
		CreatedAt: chirp.CreatedAt,
		UpdatedAt: chirp.UpdatedAt,
		Body:      chirp.Body,
		UserID:    chirp.UserID,
	}

	// If creating the record goes well, respond with a 201 status code and the full chirp resource
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(mappedChirp)
}

// HandleCreateUser creates a new user from the email address in the request body
// and returns the user's ID, email, and timestamps in the response body.
func (cfg *ApiConfig) HandleCreateUser(w http.ResponseWriter, r *http.Request) {
	// Parse the JSON body of the request into a CreateUserRequest struct
	var createUserRequest CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&createUserRequest); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Invalid request body"})
		return
	}

	// Hash the provided password
	hashedPassword, err := auth.HashPassword(createUserRequest.HashedPassword)
	if err != nil {
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Failed to hash provided password"})
		return
	}
	createUserRequest.HashedPassword = hashedPassword

	// Create a new user in the database
	user, err := cfg.DbQueries.CreateUser(r.Context(), database.CreateUserParams(createUserRequest))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Failed to create user"})
		return
	}

	// Map user to the MappedUser struct in order to control the JSON keys
	mappedUser := MappedUser{
		ID:          user.ID,
		CreatedAt:   user.CreatedAt,
		UpdatedAt:   user.UpdatedAt,
		Email:       user.Email,
		IsChirpyRed: user.IsChirpyRed,
	}

	// Respond with 200 OK and a valid response if the user was created successfully
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(mappedUser)
}

// HandleGetAllChirps retrieves all chirps from the database and returns them as a
// JSON object in the response. The query parameter "author_id" can be used to
// retrieve all chirps for the given author. The query parameter "sort" can be used
// to sort the chirps in ascending or descending order of their creation date.
// If there is an error retrieving the chirps, it responds with an appropriate error
// status and message. If the chirps are successfully retrieved, it responds with a
// 200 OK status and a valid JSON response.
func (cfg *ApiConfig) HandleGetAllChirps(w http.ResponseWriter, r *http.Request) {

	// Check if the author_id and/or the sort query parameter is provided
	authorID := r.URL.Query().Get("author_id")
	sort := r.URL.Query().Get("sort")

	var chirps []database.Chirp
	var err error
	if sort == "desc" {
		if authorID != "" {
			// Get all chirps for the given author
			chirps, err = cfg.DbQueries.GetUserChirpsDESC(r.Context(), uuid.MustParse(authorID))
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(ErrorResponse{Error: "Failed to get chirps for author"})
				return
			}
		} else {
			// Get all chirps from the database
			chirps, err = cfg.DbQueries.GetAllChirpsDESC(r.Context())
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(ErrorResponse{Error: "Failed to get all chirps"})
				return
			}
		}
	} else {
		if authorID != "" {
			// Get all chirps for the given author
			chirps, err = cfg.DbQueries.GetUserChirps(r.Context(), uuid.MustParse(authorID))
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(ErrorResponse{Error: "Failed to get chirps for author"})
				return
			}
		} else {
			// Get all chirps from the database
			chirps, err = cfg.DbQueries.GetAllChirps(r.Context())
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(ErrorResponse{Error: "Failed to get all chirps"})
				return
			}
		}
	}

	// Map the chirps to the MappedChirp struct to control the JSON keys
	var mappedChirps []MappedChirp
	for _, chirp := range chirps {
		mappedChirp := MappedChirp{
			ID:        chirp.ID,
			CreatedAt: chirp.CreatedAt,
			UpdatedAt: chirp.UpdatedAt,
			Body:      chirp.Body,
			UserID:    chirp.UserID,
		}
		mappedChirps = append(mappedChirps, mappedChirp)
	}

	// Respond with 200 OK and a valid response if successful
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	// Create a new JSON encoder and configure it for pretty printing
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ") // Set indent with two spaces
	encoder.Encode(mappedChirps)
}

// HandleGetChirp retrieves a chirp from the database by its ID and returns it
// as a JSON object in the response. It maps the database chirp record to the
// MappedChirp struct to ensure consistent JSON keys. If the chirp is found,
// it responds with a 200 OK status and a valid JSON response. If the chirp is
// not found, it responds with a 404 status and an error message.
func (cfg *ApiConfig) HandleGetChirp(w http.ResponseWriter, r *http.Request) {
	// Get the chirp ID from the path parameter and convert it to a UUID object
	pathParameter := r.PathValue("chirpID")
	chirpID, err := uuid.Parse(pathParameter)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Invalid chirp ID"})
		return
	}

	// Get the requested chirp from the database
	chirp, err := cfg.DbQueries.GetChirp(r.Context(), chirpID)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Failed to find chirp with ID: " + pathParameter})
		return
	}

	// Map the chirp struct to a MappedChirp struct to control the JSON keys
	mappedChirp := MappedChirp{
		ID:        chirp.ID,
		CreatedAt: chirp.CreatedAt,
		UpdatedAt: chirp.UpdatedAt,
		Body:      chirp.Body,
		UserID:    chirp.UserID,
	}

	// If the chirp is found, respond with a 200 OK code and the found chirp
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(mappedChirp)
}

// HandleLoginUser processes a login request by verifying the user's email
// and password. It expects a JSON request body containing the user's email
// and password, which is parsed into a LoginUserRequest struct. The function
// retrieves the user from the database using the provided email and checks
// if the password matches the stored hash. If the credentials are valid, it
// generates a JWT access token and a refresh token, adds the refresh token
// to the database, and returns a 200 OK response with the user's details,
// access token, and refresh token. If the request body is invalid, or the
// email or password is incorrect, it returns an appropriate error response.
func (cfg *ApiConfig) HandleLoginUser(w http.ResponseWriter, r *http.Request) {
	// Parse the JSON body of the request into a LoginUserRequest struct
	var loginUserRequest LoginUserRequest
	if err := json.NewDecoder(r.Body).Decode(&loginUserRequest); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Invalid request body"})
		return
	}

	// Match the user email to an email in the database
	user, err := cfg.DbQueries.GetUserByEmail(r.Context(), loginUserRequest.Email)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Incorrect email or password"})
		return
	}

	// Check to see if their password matches the stored hash
	if err := auth.CheckPasswordHash(loginUserRequest.Password, user.HashedPassword); err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Incorrect email or password"})
		return
	}

	// Set the expiration time for the access token (JWT) to 1 hour
	token, err := auth.MakeJWT(user.ID, cfg.TokenSecret, 3600*time.Second)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Failed to generate JWT"})
		return
	}
	w.Header().Set("Authorization", "Bearer "+token)

	// Create a refresh token and insert it into the database
	makeRefreshToken, err := auth.MakeRefreshToken()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Failed to generate refresh token"})
		return
	}

	createRefreshToken := database.CreateRefreshTokenParams{
		Token:  makeRefreshToken,
		UserID: user.ID,
	}

	if cfg.DbQueries.CreateRefreshToken(r.Context(), createRefreshToken) != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Failed to add refresh token to database"})
		return
	}

	// Map user to the MappedUser struct in order to control the JSON keys
	mappedUser := MappedUser{
		ID:           user.ID,
		CreatedAt:    user.CreatedAt,
		UpdatedAt:    user.UpdatedAt,
		Email:        user.Email,
		IsChirpyRed:  user.IsChirpyRed,
		Token:        token,
		RefreshToken: makeRefreshToken,
	}

	// If the email and passwords match, return a 200 OK response and a copy of the user resource with the token
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(mappedUser)
}

// HandleRefresh processes a request to refresh a user's access token using
// their refresh token. It expects the refresh token to be provided in the
// Authorization header of the request. If the token is missing or invalid,
// it returns an appropriate error response. If the token is valid, it
// generates a new access token and returns it in the response body. If there
// is an error generating the new token, it responds with a 500 status and an
// error message.
func (cfg *ApiConfig) HandleRefresh(w http.ResponseWriter, r *http.Request) {
	refreshToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Missing refresh token"})
		return
	}

	user, err := cfg.DbQueries.GetUserFromRefreshToken(r.Context(), refreshToken)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Invalid refresh token"})
		return
	}

	if user.RevokedAt.Valid {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Refresh token is expired"})
		return
	}

	token, err := auth.MakeJWT(user.ID, cfg.TokenSecret, 3600*time.Second)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Failed to generate JWT"})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(NewJWT{Token: token})
}

// HandleRevoke processes a request to revoke a user's refresh token.
// It extracts the refresh token from the Authorization header, validates it,
// and marks the associated refresh token as revoked in the database. If the
// refresh token is missing or invalid, or if there is an error storing the
// revocation, it responds with an appropriate error status and error message.
func (cfg *ApiConfig) HandleRevoke(w http.ResponseWriter, r *http.Request) {
	refreshToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Missing refresh token"})
		return
	}

	err = cfg.DbQueries.RevokeRefreshToken(r.Context(), refreshToken)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Failed to revoke refresh token"})
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// HandleUpdateUser processes a request to update a user's email and/or password.
// It expects the user to provide an access token in the Authorization header,
// and the new email and password in the request body. If the access token is
// missing or invalid, it responds with a 401 status code and an appropriate
// error message. It hashes the new password and updates the user in the
// database with the new email and hashed password. If there is an error
// storing the update, it responds with a 500 status code and an appropriate
// error message. Otherwise, it responds with a 200 status code and the newly
// updated User resource.
func (cfg *ApiConfig) HandleUpdateUser(w http.ResponseWriter, r *http.Request) {
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Missing access token"})
		return
	}

	userID, err := auth.ValidateJWT(token, cfg.TokenSecret)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Invalid access token"})
		return
	}

	var updateUserRequest UpdateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&updateUserRequest); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Invalid request body"})
		return
	}

	hashedPassword, err := auth.HashPassword(updateUserRequest.Password)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Failed to hash password"})
		return
	}

	updateUserParams := database.UpdateUserParams{
		ID:             userID,
		Email:          updateUserRequest.Email,
		HashedPassword: hashedPassword,
	}

	user, err := cfg.DbQueries.UpdateUser(r.Context(), updateUserParams)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Failed to update user"})
		return
	}

	mappedUser := MappedUser{
		ID:          user.ID,
		CreatedAt:   user.CreatedAt,
		UpdatedAt:   user.UpdatedAt,
		Email:       user.Email,
		IsChirpyRed: user.IsChirpyRed,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(mappedUser)
}

// HandleDeleteChirp deletes a chirp from the database by its ID. The function expects
// the chirp ID to be provided as a path parameter and the user's JWT to be provided
// in the Authorization header. It validates the JWT and checks if the chirp belongs
// to the authenticated user. If the chirp ID is invalid, the JWT is missing or invalid,
// or if the chirp does not belong to the user, it responds with an appropriate error
// status and message. If the chirp is successfully deleted, it responds with a 204
// No Content status.
func (cfg *ApiConfig) HandleDeleteChirp(w http.ResponseWriter, r *http.Request) {
	// Get the chirp ID from the path parameter and convert it to a UUID object
	pathParameter := r.PathValue("chirpID")
	chirpID, err := uuid.Parse(pathParameter)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Invalid chirp ID"})
		return
	}

	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Invalid authorization header"})
		return
	}

	userID, err := auth.ValidateJWT(token, cfg.TokenSecret)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Invalid JWT"})
		return
	}

	chirp, err := cfg.DbQueries.GetChirp(r.Context(), chirpID)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Chirp not found"})
		return
	}

	if chirp.UserID != userID {
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "You can only delete your own chirps"})
		return
	}

	err = cfg.DbQueries.DeleteChirp(r.Context(), chirpID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Failed to delete chirp"})
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// HandleStripeEvent handles a Stripe event and upgrades the corresponding user
// to a Chirpy Red subscription if the event is user.upgraded. It expects the
// event to be passed in the request body as a StripeEvent struct, and expects
// the Stripe API key to be passed in the Authorization header. If the event
// is invalid, it responds with a 400 status code and an appropriate error
// message. If the API key is invalid, it responds with a 403 status code and
// an appropriate error message. If the user is not found, it responds with a
// 404 status code and an appropriate error message. If the event is not
// user.upgraded, it responds with a 204 status code. Otherwise, it upgrades
// the user and responds with a 204 status code.
func (cfg *ApiConfig) HandleStripeEvent(w http.ResponseWriter, r *http.Request) {
	var stripeEvent StripeEvent
	err := json.NewDecoder(r.Body).Decode(&stripeEvent)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Invalid request body"})
		return
	}

	if stripeEvent.Event != "user.upgraded" {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	apiKey, err := auth.GetAPIKey(r.Header)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Authorization header is invalid"})
		return
	}

	if apiKey != cfg.StripeKey {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Invalid API key"})
		return
	}

	err = cfg.DbQueries.UpgradeUserToChirpyRed(r.Context(), stripeEvent.Data.UserID)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Failed to find user"})
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
