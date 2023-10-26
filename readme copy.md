## Ribbit Fleet Manager
Current functionality is user and device associations. As this is an open source project, we gladly welcome any new developers to help us expand current functinoality. This app queries the Golioth REST APIs (reference docs available on their [website](https://docs.golioth.io/reference))

## Getting Started
To run the app locally, you'll need a MySQL server and you'll need to configure the following three environment variables:
* "DB_CONN"
* "GOLIOTH_PROJECT_ID"
* "GOLIOTH_API_KEY"

## Ribbit REST APIs
* `POST /signup/:email/:password`
* `POST /signin/:email/:password`
* `POST /createNewDevice`
* `DELETE /users/:email`
* For developer testing only: `POST /createDeviceGolioth`
    * It creates a device directly in the golioth account without associating it to a user