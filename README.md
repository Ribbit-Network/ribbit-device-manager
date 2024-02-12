## Ribbit Device Manager
Current functionality is user and device associations. As this is an open source project, we gladly welcome any new developers to help us expand current functinoality. This app queries the Golioth REST APIs (reference docs available on their [website](https://docs.golioth.io/reference))

## Signing up for ribbit account
If the app is being run locally, below are a list of the URLs that are provided for the user to be able to signup, login and create a new device:
* http://localhost:8080 : the signup page
* http://localhost:8080/login : the login page. Once logged in, you will be direct to a page that will allow you to add devices

# For Developers
This section is intended to provide enough information for you to be able to run the app in your local environment.

## Getting Started
To run the app locally, you'll need a MySQL server and you'll need to configure the following three environment variables:
* "DSN_ENV" : Specific to your local db creds
* "GOLIOTH_PROJECT_ID" : available in the Golioth user terminal
* "GOLIOTH_API_KEY": available in the Golioth user terminal

If you are using vscode. Below is a sample json, the provided environment variables will not work and have to be replaced with ones that are specific to your setup.
```
{
    "version": "0.2.0",
    "configurations": [
        {
            "name": "ribbit",
            "type": "go",
            "request": "launch",
            "mode": "auto",
            "program": "${workspaceFolder}/cmd/main.go",
            "env" :
            {
                "DSN_ENV" : "admin:password123@tcp(127.0.0.1:3306)/ribbit?charset=utf8mb4&parseTime=True&loc=Local",
                "GOLIOTH_PROJECT_ID" : "ribbit-test-123456",
                "GOLIOTH_API_KEY" : "Z2s6Kf1w3FqRTVjg5My9ZpXtDhNJxQ8P",
            }
        }
    ]
}
```
## Browser Access
Since we don't have a homepage dashboard yet, below are the two webpages that you will need to be able to access all respective functionality. The remaining info will be presented to you on these pages.
* `localhost:8080/` : signup page
* `localhost:8080/login` : login page


## Ribbit REST APIs
The APIs available are listed below and can be accessed by using a separate API tool such as postman.

* `GET /` : loads signup user page
* `POST /signup/:email/:password` : signup a new user
* `GET /login` : load the login user page
* `GET /logout` : load the logout user page
* `POST /login/:email/:password` : sign in user assuming the signup is complete 
* `DELETE /users/:email` : deletes the respective user email
* `POST /createNewDevice` : creates a device in the golioth system and associates it with the specified user email

## OPEN ITEMS
* server user a list of golioth devices associated to their user account and present on the UI on the `loggedin.html` and `devices.html` page
* Improve UX. Make the UI a lot nicer and improve navigation
* Create an admin user page
* TODOs throughout the code. There are a few of these