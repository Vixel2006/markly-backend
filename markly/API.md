# Markly Backend API Documentation

This document provides a comprehensive overview of the Markly backend API, designed to assist frontend developers in integrating with the service.

## Base URL

The base URL for all API endpoints is dependent on your deployment. During local development, it's typically `http://localhost:8080` (or the port configured in your `.env` file).

## Authentication

Markly API uses JSON Web Tokens (JWT) for authentication.

### How to Authenticate

1.  **Register or Log In:** Obtain a JWT by registering a new user or logging in with existing credentials via the `/api/auth/register` or `/api/auth/login` endpoints.
2.  **Receive Token:** Upon successful registration or login, the API will return a JWT.
3.  **Include Token in Requests:** For all protected routes, include the JWT in the `Authorization` header of your HTTP requests in the format `Bearer <YOUR_JWT_TOKEN>`.

    Example:
    ```
    Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
    ```

## Common Response Structures

### Success Response

Successful requests typically return a `200 OK` status code (or `201 Created` for resource creation, `204 No Content` for successful deletion without a response body) and a JSON object or array containing the requested data.

Example (200 OK):
```json
{
  "id": "654321098765432109876543",
  "name": "My Category",
  "emoji": "📚"
}
```

### Error Response

Error responses typically return an appropriate HTTP status code (e.g., `400 Bad Request`, `401 Unauthorized`, `404 Not Found`, `409 Conflict`, `500 Internal Server Error`) and a JSON object with an `error` field describing the issue.

Example (400 Bad Request):
```json
{
  "error": "Invalid JSON input: unexpected end of JSON input"
}
```

Example (401 Unauthorized):
```json
{
  "error": "Missing token"
}
```

Example (409 Conflict):
```json
{
  "error": "Email already exists"
}
```

---

## API Endpoints

### 1. Common Endpoints

#### 1.1. Get Hello World Message

*   **URL:** `/`
*   **Method:** `GET`
*   **Description:** Returns a simple "Hello World" message.
*   **Authentication:** None
*   **Success Response (200 OK):**
    ```json
    {
      "message": "Hello World"
    }
    ```

#### 1.2. Get Health Status

*   **URL:** `/health`
*   **Method:** `GET`
*   **Description:** Checks the health of the database connection.
*   **Authentication:** None
*   **Success Response (200 OK):**
    ```json
    {
      "message": "It's healthy"
    }
    ```

---

### 2. Authentication Endpoints

#### 2.1. Register User

*   **URL:** `/api/auth/register`
*   **Method:** `POST`
*   **Description:** Registers a new user.
*   **Authentication:** None
*   **Request Body:** `application/json`
    ```json
    {
      "username": "john_doe",
      "email": "john.doe@example.com",
      "password": "securepassword123"
    }
    ```
    *   `username` (string, required): The user's chosen username.
    *   `email` (string, required): The user's email address (must be unique).
    *   `password` (string, required): The user's password.
*   **Success Response (201 Created):**
    ```json
    {
      "id": "654321098765432109876543",
      "username": "john_doe",
      "email": "john.doe@example.com"
    }
    ```
    *   `id` (string): The unique ID of the newly created user.
    *   `username` (string): The registered username.
    *   `email` (string): The registered email.
*   **Error Responses:**
    *   `400 Bad Request`: Invalid JSON or missing required fields.
    *   `409 Conflict`: Email already exists.
    *   `500 Internal Server Error`: Failed to hash password or create user.

#### 2.2. Login User

*   **URL:** `/api/auth/login`
*   **Method:** `POST`
*   **Description:** Logs in an existing user and returns a JWT.
*   **Authentication:** None
*   **Request Body:** `application/json`
    ```json
    {
      "email": "john.doe@example.com",
      "password": "securepassword123"
    }
    ```
    *   `email` (string, required): The user's email address.
    *   `password` (string, required): The user's password.
*   **Success Response (200 OK):**
    ```json
    {
      "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
    }
    ```
    *   `token` (string): The JWT for authenticated requests.
*   **Error Responses:**
    *   `400 Bad Request`: Invalid request body.
    *   `401 Unauthorized`: Invalid credentials.
    *   `500 Internal Server Error`: Failed to generate token.

#### 2.3. Get My Profile

*   **URL:** `/api/me`
*   **Method:** `GET`
*   **Description:** Retrieves the authenticated user's profile information.
*   **Authentication:** Required (JWT)
*   **Success Response (200 OK):**
    ```json
    {
      "id": "654321098765432109876543",
      "username": "john_doe",
      "email": "john.doe@example.com"
    }
    ```
    *   `id` (string): The user's unique ID.
    *   `username` (string): The user's username.
    *   `email` (string): The user's email.
*   **Error Responses:**
    *   `401 Unauthorized`: Missing or invalid token.
    *   `404 Not Found`: User not found.
    *   `500 Internal Server Error`: Failed to fetch user profile.

#### 2.4. Update My Profile

*   **URL:** `/api/me`
*   **Method:** `PATCH` or `PUT`
*   **Description:** Updates the authenticated user's profile information.
*   **Authentication:** Required (JWT)
*   **Request Body:** `application/json`
    ```json
    {
      "username": "new_john_doe",
      "email": "new.john.doe@example.com",
      "password": "newsecurepassword123"
    }
    ```
    *   `username` (string, optional): New username.
    *   `email` (string, optional): New email address (must be unique).
    *   `password` (string, optional): New password.
*   **Success Response (200 OK):**
    ```json
    {
      "id": "654321098765432109876543",
      "username": "new_john_doe",
      "email": "new.john.doe@example.com"
    }
    ```
    *   Returns the updated user object (without password).
*   **Error Responses:**
    *   `400 Bad Request`: Invalid JSON payload or no valid fields for update.
    *   `401 Unauthorized`: Missing or invalid token.
    *   `404 Not Found`: User not found.
    *   `500 Internal Server Error`: Failed to update profile or email already in use.

#### 2.5. Delete My Profile

*   **URL:** `/api/me`
*   **Method:** `DELETE`
*   **Description:** Deletes the authenticated user's account.
*   **Authentication:** Required (JWT)
*   **Success Response (204 No Content):**
    *   No response body.
*   **Error Responses:**
    *   `401 Unauthorized`: Missing or invalid token.
    *   `404 Not Found`: User not found.
    *   `500 Internal Server Error`: Failed to delete account.

---

### 3. Bookmark Endpoints

#### 3.1. Get All Bookmarks

*   **URL:** `/api/bookmarks`
*   **Method:** `GET`
*   **Description:** Retrieves all bookmarks for the authenticated user, with optional filtering.
*   **Authentication:** Required (JWT)
*   **Query Parameters (Optional):**
    *   `tags` (string): Comma-separated list of tag ObjectIDs to filter by.
    *   `category` (string): Category ObjectID to filter by.
    *   `collections` (string): Comma-separated list of collection ObjectIDs to filter by.
    *   `isFav` (boolean): `true` to get favorite bookmarks, `false` for non-favorites.
*   **Success Response (200 OK):**
    ```json
    [
      {
        "id": "654321098765432109876543",
        "user_id": "654321098765432109876543",
        "url": "https://example.com/bookmark1",
        "title": "My First Bookmark",
        "summary": "A brief summary of the first bookmark.",
        "tags": ["654321098765432109876544"],
        "collections": ["654321098765432109876545"],
        "category": "654321098765432109876546",
        "is_fav": true,
        "created_at": "2023-11-17T10:00:00Z"
      }
    ]
    ```
    *   Returns an array of `Bookmark` objects.
*   **Error Responses:**
    *   `400 Bad Request`: Invalid query parameter format.
    *   `401 Unauthorized`: Missing or invalid token.
    *   `500 Internal Server Error`: Failed to retrieve bookmarks.

#### 3.2. Add New Bookmark

*   **URL:** `/api/bookmarks`
*   **Method:** `POST`
*   **Description:** Adds a new bookmark for the authenticated user.
*   **Authentication:** Required (JWT)
*   **Request Body:** `application/json`
    ```json
    {
      "url": "https://example.com/new-bookmark",
      "title": "A New Interesting Article",
      "summary": "This is a summary of the new article.",
      "tags": ["654321098765432109876544", "654321098765432109876547"],
      "collections": ["654321098765432109876545"],
      "category_id": "654321098765432109876546",
      "is_fav": false
    }
    ```
    *   `url` (string, required): The URL of the bookmark.
    *   `title` (string, required): The title of the bookmark.
    *   `summary` (string, optional): A summary of the bookmark.
    *   `tags` (array of strings, optional): Array of Tag ObjectIDs.
    *   `collections` (array of strings, optional): Array of Collection ObjectIDs.
    *   `category_id` (string, optional): Category ObjectID.
    *   `is_fav` (boolean, required): Whether the bookmark is a favorite.
*   **Success Response (201 Created):**
    ```json
    {
      "id": "654321098765432109876548",
      "user_id": "654321098765432109876543",
      "url": "https://example.com/new-bookmark",
      "title": "A New Interesting Article",
      "summary": "This is a summary of the new article.",
      "tags": ["654321098765432109876544", "654321098765432109876547"],
      "collections": ["654321098765432109876545"],
      "category": "654321098765432109876546",
      "is_fav": false,
      "created_at": "2023-11-17T10:05:00Z"
    }
    ```
    *   Returns the newly created `Bookmark` object.
*   **Error Responses:**
    *   `400 Bad Request`: Invalid JSON, missing required fields, or invalid reference IDs (tags, collections, category).
    *   `401 Unauthorized`: Missing or invalid token.
    *   `500 Internal Server Error`: Failed to add bookmark.

#### 3.3. Get Bookmark by ID

*   **URL:** `/api/bookmarks/{id}`
*   **Method:** `GET`
*   **Description:** Retrieves a specific bookmark by its ID for the authenticated user.
*   **Authentication:** Required (JWT)
*   **URL Parameters:**
    *   `id` (string, required): The ObjectID of the bookmark.
*   **Success Response (200 OK):**
    ```json
    {
      "id": "654321098765432109876543",
      "user_id": "654321098765432109876543",
      "url": "https://example.com/bookmark1",
      "title": "My First Bookmark",
      "summary": "A brief summary of the first bookmark.",
      "tags": ["654321098765432109876544"],
      "collections": ["654321098765432109876545"],
      "category": "654321098765432109876546",
      "is_fav": true,
      "created_at": "2023-11-17T10:00:00Z"
    }
    ```
    *   Returns the `Bookmark` object.
*   **Error Responses:**
    *   `400 Bad Request`: Invalid ID format.
    *   `401 Unauthorized`: Missing or invalid token.
    *   `404 Not Found`: Bookmark not found or does not belong to the user.
    *   `500 Internal Server Error`: Failed to retrieve bookmark.

#### 3.4. Delete Bookmark

*   **URL:** `/api/bookmarks/{id}`
*   **Method:** `DELETE`
*   **Description:** Deletes a specific bookmark by its ID for the authenticated user.
*   **Authentication:** Required (JWT)
*   **URL Parameters:**
    *   `id` (string, required): The ObjectID of the bookmark.
*   **Success Response (204 No Content):**
    *   No response body.
*   **Error Responses:**
    *   `400 Bad Request`: Invalid ID format.
    *   `401 Unauthorized`: Missing or invalid token.
    *   `404 Not Found`: Bookmark not found or not authorized to delete.
    *   `500 Internal Server Error`: Failed to delete bookmark.

#### 3.5. Update Bookmark

*   **URL:** `/api/bookmarks/{id}`
*   **Method:** `PUT`
*   **Description:** Updates an existing bookmark for the authenticated user.
*   **Authentication:** Required (JWT)
*   **URL Parameters:**
    *   `id` (string, required): The ObjectID of the bookmark.
*   **Request Body:** `application/json`
    ```json
    {
      "url": "https://example.com/updated-bookmark",
      "title": "Updated Article Title",
      "summary": "An updated summary.",
      "tags": ["654321098765432109876544"],
      "collections": [],
      "category_id": null,
      "is_fav": true
    }
    ```
    *   `url` (string, optional): New URL.
    *   `title` (string, optional): New title.
    *   `summary` (string, optional): New summary.
    *   `tags` (array of strings, optional): New array of Tag ObjectIDs.
    *   `collections` (array of strings, optional): New array of Collection ObjectIDs.
    *   `category_id` (string or null, optional): New Category ObjectID, or `null` to clear.
    *   `is_fav` (boolean, optional): New favorite status.
*   **Success Response (200 OK):**
    ```json
    {
      "id": "654321098765432109876543",
      "user_id": "654321098765432109876543",
      "url": "https://example.com/updated-bookmark",
      "title": "Updated Article Title",
      "summary": "An updated summary.",
      "tags": ["654321098765432109876544"],
      "collections": [],
      "category": null,
      "is_fav": true,
      "created_at": "2023-11-17T10:00:00Z"
    }
    ```
    *   Returns the updated `Bookmark` object.
*   **Error Responses:**
    *   `400 Bad Request`: Invalid JSON, invalid ID format, no valid fields for update, or invalid reference IDs.
    *   `401 Unauthorized`: Missing or invalid token.
    *   `404 Not Found`: Bookmark not found or not authorized to update.
    *   `500 Internal Server Error`: Failed to update bookmark.

---

### 4. Category Endpoints

#### 4.1. Add New Category

*   **URL:** `/api/categories`
*   **Method:** `POST`
*   **Description:** Adds a new category for the authenticated user.
*   **Authentication:** Required (JWT)
*   **Request Body:** `application/json`
    ```json
    {
      "name": "Technology",
      "emoji": "💻"
    }
    ```
    *   `name` (string, required): The name of the category (must be unique per user).
    *   `emoji` (string, optional): An emoji associated with the category.
*   **Success Response (201 Created):**
    ```json
    {
      "id": "654321098765432109876549",
      "user_id": "654321098765432109876543",
      "name": "Technology",
      "emoji": "💻"
    }
    ```
    *   Returns the newly created `Category` object.
*   **Error Responses:**
    *   `400 Bad Request`: Invalid JSON.
    *   `401 Unauthorized`: Missing or invalid token.
    *   `409 Conflict`: Category name already exists for this user.
    *   `500 Internal Server Error`: Failed to insert category.

#### 4.2. Get All Categories

*   **URL:** `/api/categories`
*   **Method:** `GET`
*   **Description:** Retrieves all categories for the authenticated user.
*   **Authentication:** Required (JWT)
*   **Success Response (200 OK):**
    ```json
    [
      {
        "id": "654321098765432109876549",
        "user_id": "654321098765432109876543",
        "name": "Technology",
        "emoji": "💻"
      },
      {
        "id": "654321098765432109876550",
        "user_id": "654321098765432109876543",
        "name": "Science",
        "emoji": "🔬"
      }
    ]
    ```
    *   Returns an array of `Category` objects.
*   **Error Responses:**
    *   `401 Unauthorized`: Missing or invalid token.
    *   `500 Internal Server Error`: Failed to fetch categories.

#### 4.3. Get Category by ID

*   **URL:** `/api/categories/{id}`
*   **Method:** `GET`
*   **Description:** Retrieves a specific category by its ID for the authenticated user.
*   **Authentication:** Required (JWT)
*   **URL Parameters:**
    *   `id` (string, required): The ObjectID of the category.
*   **Success Response (200 OK):**
    ```json
    {
      "id": "654321098765432109876549",
      "user_id": "654321098765432109876543",
      "name": "Technology",
      "emoji": "💻"
    }
    ```
    *   Returns the `Category` object.
*   **Error Responses:**
    *   `400 Bad Request`: Invalid ID format.
    *   `401 Unauthorized`: Missing or invalid token.
    *   `404 Not Found`: Category not found or does not belong to the user.
    *   `500 Internal Server Error`: Failed to retrieve category.

#### 4.4. Delete Category

*   **URL:** `/api/categories/{id}`
*   **Method:** `DELETE`
*   **Description:** Deletes a specific category by its ID for the authenticated user.
*   **Authentication:** Required (JWT)
*   **URL Parameters:**
    *   `id` (string, required): The ObjectID of the category.
*   **Success Response (200 OK):**
    ```json
    {
      "message": "Category deleted successfully",
      "deleted_count": 1
    }
    ```
*   **Error Responses:**
    *   `400 Bad Request`: Invalid ID format.
    *   `401 Unauthorized`: Missing or invalid token.
    *   `404 Not Found`: Category not found or unauthorized.
    *   `500 Internal Server Error`: Failed to delete category.

#### 4.5. Update Category

*   **URL:** `/api/categories/{id}`
*   **Method:** `PUT`
*   **Description:** Updates an existing category for the authenticated user.
*   **Authentication:** Required (JWT)
*   **URL Parameters:**
    *   `id` (string, required): The ObjectID of the category.
*   **Request Body:** `application/json`
    ```json
    {
      "name": "Updated Technology",
      "emoji": "⚙️"
    }
    ```
    *   `name` (string, optional): New name for the category.
    *   `emoji` (string, optional): New emoji for the category.
*   **Success Response (200 OK):**
    ```json
    {
      "id": "654321098765432109876549",
      "user_id": "654321098765432109876543",
      "name": "Updated Technology",
      "emoji": "⚙️"
    }
    ```
    *   Returns the updated `Category` object.
*   **Error Responses:**
    *   `400 Bad Request`: Invalid JSON payload or no fields to update.
    *   `401 Unauthorized`: Missing or invalid token.
    *   `404 Not Found`: Category not found or unauthorized.
    *   `409 Conflict`: Category name already exists for this user.
    *   `500 Internal Server Error`: Failed to update category.

---

### 5. Collection Endpoints

#### 5.1. Add New Collection

*   **URL:** `/api/collections`
*   **Method:** `POST`
*   **Description:** Adds a new collection for the authenticated user.
*   **Authentication:** Required (JWT)
*   **Request Body:** `application/json`
    ```json
    {
      "name": "My Reading List"
    }
    ```
    *   `name` (string, required): The name of the collection (must be unique per user).
*   **Success Response (201 Created):**
    ```json
    {
      "id": "654321098765432109876551",
      "user_id": "654321098765432109876543",
      "name": "My Reading List"
    }
    ```
    *   Returns the newly created `Collection` object.
*   **Error Responses:**
    *   `400 Bad Request`: Invalid JSON.
    *   `401 Unauthorized`: Missing or invalid token.
    *   `409 Conflict`: Collection name already exists for this user.
    *   `500 Internal Server Error`: Failed to insert collection.

#### 5.2. Get All Collections

*   **URL:** `/api/collections`
*   **Method:** `GET`
*   **Description:** Retrieves all collections for the authenticated user.
*   **Authentication:** Required (JWT)
*   **Success Response (200 OK):**
    ```json
    [
      {
        "id": "654321098765432109876551",
        "user_id": "654321098765432109876543",
        "name": "My Reading List"
      },
      {
        "id": "654321098765432109876552",
        "user_id": "654321098765432109876543",
        "name": "Work Resources"
      }
    ]
    ```
    *   Returns an array of `Collection` objects.
*   **Error Responses:**
    *   `401 Unauthorized`: Missing or invalid token.
    *   `500 Internal Server Error`: Failed to fetch collections.

#### 5.3. Get Collection by ID

*   **URL:** `/api/collections/{id}`
*   **Method:** `GET`
*   **Description:** Retrieves a specific collection by its ID for the authenticated user.
*   **Authentication:** Required (JWT)
*   **URL Parameters:**
    *   `id` (string, required): The ObjectID of the collection.
*   **Success Response (200 OK):**
    ```json
    {
      "id": "654321098765432109876551",
      "user_id": "654321098765432109876543",
      "name": "My Reading List"
    }
    ```
    *   Returns the `Collection` object.
*   **Error Responses:**
    *   `400 Bad Request`: Invalid ID format.
    *   `401 Unauthorized`: Missing or invalid token.
    *   `404 Not Found`: Collection not found or does not belong to the user.
    *   `500 Internal Server Error`: Failed to retrieve collection.

#### 5.4. Delete Collection

*   **URL:** `/api/collections/{id}`
*   **Method:** `DELETE`
*   **Description:** Deletes a specific collection by its ID for the authenticated user.
*   **Authentication:** Required (JWT)
*   **URL Parameters:**
    *   `id` (string, required): The ObjectID of the collection.
*   **Success Response (200 OK):**
    ```json
    {
      "message": "Collection deleted successfully",
      "deleted_count": 1
    }
    ```
*   **Error Responses:**
    *   `400 Bad Request`: Invalid ID format.
    *   `401 Unauthorized`: Missing or invalid token.
    *   `404 Not Found`: Collection not found or unauthorized.
    *   `500 Internal Server Error`: Failed to delete collection.

#### 5.5. Update Collection

*   **URL:** `/api/collections/{id}`
*   **Method:** `PUT`
*   **Description:** Updates an existing collection for the authenticated user.
*   **Authentication:** Required (JWT)
*   **URL Parameters:**
    *   `id` (string, required): The ObjectID of the collection.
*   **Request Body:** `application/json`
    ```json
    {
      "name": "My Updated Reading List"
    }
    ```
    *   `name` (string, optional): New name for the collection.
*   **Success Response (200 OK):**
    ```json
    {
      "id": "654321098765432109876551",
      "user_id": "654321098765432109876543",
      "name": "My Updated Reading List"
    }
    ```
    *   Returns the updated `Collection` object.
*   **Error Responses:**
    *   `400 Bad Request`: Invalid JSON payload or no fields to update.
    *   `401 Unauthorized`: Missing or invalid token.
    *   `404 Not Found`: Collection not found or unauthorized.
    *   `409 Conflict`: Collection name already exists for this user.
    *   `500 Internal Server Error`: Failed to update collection.

---

### 6. Tag Endpoints

#### 6.1. Add New Tag

*   **URL:** `/api/tags`
*   **Method:** `POST`
*   **Description:** Adds a new tag for the authenticated user.
*   **Authentication:** Required (JWT)
*   **Request Body:** `application/json`
    ```json
    {
      "name": "GoLang"
    }
    ```
    *   `name` (string, required): The name of the tag (must be unique per user).
*   **Success Response (201 Created):**
    ```json
    {
      "id": "654321098765432109876553",
      "name": "GoLang",
      "user_id": "654321098765432109876543",
      "weeklyCount": 0,
      "prevCount": 0,
      "createdAt": "2023-11-17T10:10:00Z"
    }
    ```
    *   Returns the newly created `Tag` object.
*   **Error Responses:**
    *   `400 Bad Request`: Invalid JSON.
    *   `401 Unauthorized`: Missing or invalid token.
    *   `409 Conflict`: Tag name already exists for this user.
    *   `500 Internal Server Error`: Failed to insert tag.

#### 6.2. Get Tags by ID

*   **URL:** `/api/tags`
*   **Method:** `GET`
*   **Description:** Retrieves specific tags by their IDs for the authenticated user.
*   **Authentication:** Required (JWT)
*   **Query Parameters:**
    *   `id` (string, required): Comma-separated list of tag ObjectIDs.
*   **Success Response (200 OK):**
    ```json
    [
      {
        "id": "654321098765432109876553",
        "name": "GoLang",
        "user_id": "654321098765432109876543",
        "weeklyCount": 0,
        "prevCount": 0,
        "createdAt": "2023-11-17T10:10:00Z"
      },
      {
        "id": "654321098765432109876554",
        "name": "MongoDB",
        "user_id": "654321098765432109876543",
        "weeklyCount": 5,
        "prevCount": 3,
        "createdAt": "2023-11-10T09:00:00Z"
      }
    ]
    ```
    *   Returns an array of `Tag` objects.
*   **Error Responses:**
    *   `400 Bad Request`: Invalid ID format.
    *   `401 Unauthorized`: Missing or invalid token.
    *   `500 Internal Server Error`: Failed to retrieve tags.

#### 6.3. Get All User Tags

*   **URL:** `/api/tags/user`
*   **Method:** `GET`
*   **Description:** Retrieves all tags belonging to the authenticated user.
*   **Authentication:** Required (JWT)
*   **Success Response (200 OK):**
    ```json
    [
      {
        "id": "654321098765432109876553",
        "name": "GoLang",
        "user_id": "654321098765432109876543",
        "weeklyCount": 0,
        "prevCount": 0,
        "createdAt": "2023-11-17T10:10:00Z"
      },
      {
        "id": "654321098765432109876554",
        "name": "MongoDB",
        "user_id": "654321098765432109876543",
        "weeklyCount": 5,
        "prevCount": 3,
        "createdAt": "2023-11-10T09:00:00Z"
      }
    ]
    ```
    *   Returns an array of `Tag` objects.
*   **Error Responses:**
    *   `401 Unauthorized`: Missing or invalid token.
    *   `500 Internal Server Error`: Failed to retrieve tags.

#### 6.4. Delete Tag

*   **URL:** `/api/tags/{id}`
*   **Method:** `DELETE`
*   **Description:** Deletes a specific tag by its ID for the authenticated user.
*   **Authentication:** Required (JWT)
*   **URL Parameters:**
    *   `id` (string, required): The ObjectID of the tag.
*   **Success Response (200 OK):**
    ```json
    {
      "message": "Tag deleted successfully",
      "deleted_count": 1
    }
    ```
*   **Error Responses:**
    *   `400 Bad Request`: Invalid ID format.
    *   `401 Unauthorized`: Missing or invalid token.
    *   `404 Not Found`: Tag not found or unauthorized.
    *   `500 Internal Server Error`: Failed to delete tag.

#### 6.5. Update Tag

*   **URL:** `/api/tags/{id}`
*   **Method:** `PUT`
*   **Description:** Updates an existing tag for the authenticated user.
*   **Authentication:** Required (JWT)
*   **URL Parameters:**
    *   `id` (string, required): The ObjectID of the tag.
*   **Request Body:** `application/json`
    ```json
    {
      "name": "Go",
      "weeklyCount": 10,
      "prevCount": 5
    }
    ```
    *   `name` (string, optional): New name for the tag.
    *   `weeklyCount` (integer, optional): New weekly count for the tag.
    *   `prevCount` (integer, optional): New previous count for the tag.
*   **Success Response (200 OK):**
    ```json
    {
      "id": "654321098765432109876553",
      "name": "Go",
      "user_id": "654321098765432109876543",
      "weeklyCount": 10,
      "prevCount": 5,
      "createdAt": "2023-11-17T10:10:00Z"
    }
    ```
    *   Returns the updated `Tag` object.
*   **Error Responses:**
    *   `400 Bad Request`: Invalid JSON payload or no fields to update.
    *   `401 Unauthorized`: Missing or invalid token.
    *   `404 Not Found`: Tag not found or unauthorized.
    *   `409 Conflict`: Tag name already exists for this user.
    *   `500 Internal Server Error`: Failed to update tag.

---

### 7. Agent Endpoints

#### 7.1. Generate Bookmark Summary

*   **URL:** `/api/agent/summarize/{id}`
*   **Method:** `POST`
*   **Description:** Generates a summary for a specific bookmark using an AI agent.
*   **Authentication:** Required (JWT)
*   **URL Parameters:**
    *   `id` (string, required): The ObjectID of the bookmark to summarize.
*   **Success Response (200 OK):**
    ```json
    {
      "id": "654321098765432109876543",
      "user_id": "654321098765432109876543",
      "url": "https://example.com/bookmark1",
      "title": "My First Bookmark",
      "summary": "This is the newly generated AI summary for the bookmark.",
      "tags": ["654321098765432109876544"],
      "collections": ["654321098765432109876545"],
      "category": "654321098765432109876546",
      "is_fav": true,
      "created_at": "2023-11-17T10:00:00Z"
    }
    ```
    *   Returns the updated `Bookmark` object with the new summary.
*   **Error Responses:**
    *   `400 Bad Request`: Invalid ID format.
    *   `401 Unauthorized`: Missing or invalid token.
    *   `404 Not Found`: Bookmark not found.
    *   `500 Internal Server Error`: Failed to generate or save summary.

#### 7.2. Summarize URL

*   **URL:** `/api/agent/summarize-url`
*   **Method:** `POST`
*   **Description:** Generates a summary for a given URL and title using an AI agent, without saving it as a bookmark.
*   **Authentication:** Required (JWT)
*   **Request Body:** `application/json`
    ```json
    {
      "url": "https://example.com/article-to-summarize",
      "title": "Article Title"
    }
    ```
    *   `url` (string, required): The URL to summarize.
    *   `title` (string, required): The title of the URL content.
*   **Success Response (200 OK):**
    ```json
    {
      "summary": "This is the AI-generated summary of the provided URL."
    }
    ```
    *   `summary` (string): The AI-generated summary.
*   **Error Responses:**
    *   `400 Bad Request`: Invalid JSON or missing URL/title.
    *   `401 Unauthorized`: Missing or invalid token.
    *   `500 Internal Server Error`: Failed to generate summary.

#### 7.4. Generate AI Suggestions

*   **URL:** `/api/agent/suggestions`
*   **Method:** `GET`
*   **Description:** Generates AI-powered bookmark suggestions based on the user's recent bookmarking activity.
*   **Authentication:** Required (JWT)
*   **Query Parameters (Optional):
    *   `bookmarks` (string): Comma-separated list of bookmark ObjectIDs to filter by.
    *   `category` (string): Category ObjectID to filter by.
    *   `collection` (string): Comma-separated list of collection ObjectIDs to filter by.
    *   `tag` (string): Comma-separated list of tag ObjectIDs to filter by.
*   **Success Response (200 OK):**
    ```json
    [
      {
        "url": "https://suggestion.com/article1",
        "title": "Suggested Article One",
        "summary": "A summary of a suggested article.",
        "category": "Technology",
        "collection": "Reading List",
        "tags": ["AI", "Future"]
      },
      {
        "url": "https://suggestion.com/article2",
        "title": "Suggested Article Two",
        "summary": "Another summary of a suggested article.",
        "category": "Science",
        "collection": "Research",
        "tags": ["Physics"]
      },
      {
        "url": "https://suggestion.com/article3",
        "title": "Suggested Article Three",
        "summary": "A third summary of a suggested article.",
        "category": "History",
        "collection": "Learning",
        "tags": ["Ancient"]
      }
    ]
    ```
    *   Returns an array of `AISuggestion` objects.
*   **Error Responses:**
    *   `401 Unauthorized`: Missing or invalid token.
    *   `500 Internal Server Error`: Failed to generate AI suggestions.
    *   `200 OK` with error message: "No recent bookmarks found to generate suggestions from. Please add some bookmarks first." (This is a specific case handled by the backend, returning 200 OK but with an informative message if no recent bookmarks are available).

#### 7.3. Generate AI Suggestions

*   **URL:** `/api/agent/suggestions`
*   **Method:** `GET`
*   **Description:** Generates AI-powered bookmark suggestions based on the user's recent bookmarking activity.
*   **Authentication:** Required (JWT)
*   **Success Response (200 OK):**
    ```json
    [
      {
        "url": "https://suggestion.com/article1",
        "title": "Suggested Article One",
        "summary": "A summary of a suggested article.",
        "category": "Technology",
        "collection": "Reading List",
        "tags": ["AI", "Future"]
      },
      {
        "url": "https://suggestion.com/article2",
        "title": "Suggested Article Two",
        "summary": "Another summary of a suggested article.",
        "category": "Science",
        "collection": "Research",
        "tags": ["Physics"]
      },
      {
        "url": "https://suggestion.com/article3",
        "title": "Suggested Article Three",
        "summary": "A third summary of a suggested article.",
        "category": "History",
        "collection": "Learning",
        "tags": ["Ancient"]
      }
    ]
    ```
    *   Returns an array of `AISuggestion` objects.
*   **Error Responses:**
    *   `401 Unauthorized`: Missing or invalid token.
    *   `500 Internal Server Error`: Failed to generate AI suggestions.
    *   `200 OK` with error message: "No recent bookmarks found to generate suggestions from. Please add some bookmarks first." (This is a specific case handled by the backend, returning 200 OK but with an informative message if no recent bookmarks are available).
