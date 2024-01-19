# Microsoft SSO Verification Service

This is a simple Go application that provides a web service for verifying user identity using Microsoft Single Sign-On (SSO). The application is designed to receive an authorization code from the Microsoft identity platform and exchange it for an access token. It then uses the access token to retrieve user profile information from Microsoft Graph.

## Prerequisites

Before running the application, make sure you have the following:

- Go installed on your system.
- A valid `config/config.json` file containing the required configuration parameters.
- Environment variables `SSO_STATE` and `SSO_CLIENT_SECRET` set for security.

## Configuration

Ensure that the `config/config.json` file is correctly configured with the following parameters:

- `client_id`: Your Microsoft application client ID.
- `redirect_uri`: The redirect URI configured in your Microsoft application.
- `grant_type`: The grant type for authentication.
- `scope`: The scope of the access requested.
- `medewerker_email`: The email address pattern for authorized employees.
- `medewerker_email2`: An additional email address pattern for authorized employees.

## Usage

1. Run the application:

    ```bash
    go run main.go
    ```

2. Access the verification endpoint:

    Open a web browser and navigate to `http://localhost/verify?code=<authorization_code>&state=<state_value>`, replacing `<authorization_code>` and `<state_value>` with the actual values received from the Microsoft authentication process.

## Endpoint

- `/verify`: Endpoint for verifying Microsoft SSO. Expects query parameters `code` and `state` for the authorization code and state value, respectively.

## Security

Make sure to set the `SSO_STATE` and `SSO_CLIENT_SECRET` environment variables for enhanced security. These values should match the state value and client secret used during the Microsoft authentication process.

## Dependencies

The application uses the following external libraries:

- `net/http`: For handling HTTP requests.
- `encoding/json`: For JSON encoding and decoding.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.