# traktshow

A command-line tool to interact with Trakt.tv.

## Features

- View your Trakt.tv watch history.
- Check your watch progress for shows.
- Get your Trakt.tv stats.

## Installation

1.  **Get Trakt.tv API Credentials:**
    - Go to the [Trakt.tv API documentation](https://trakt.tv/oauth/applications) and create a new application.
    - You will get a `Client ID` and `Client Secret`.

2.  **Configure traktshow:**
    ```bash
    go install github.com/zm/traktshow
    traktshow config --client-id YOUR_CLIENT_ID --client-secret YOUR_CLIENT_SECRET
    ```

3.  **Login:**
    ```bash
    traktshow login
    ```
    This will open a browser window for you to authorize the application.

## Usage

- **View History:**
  ```bash
  traktshow history
  ```

- **View Progress:**
  ```bash
  traktshow progress
  ```

- **View Stats:**
  ```bash
  traktshow stats
  ```