# WordPresser

Utility to query the stats of all of your WordPress sites from the command line. Needs a `.env` file with the following settings:

```
REDIRECT_URI=https://your.redirect.site
CLIENT_ID=12345
CLIENT_SECRET=BiGL0nGSecRetStr1ng
```

On successful authentication, program drops a cache file of your OAuth2 access token. To reauthenticate, delete the `.token` file.

## Screenshot
![WordPresser In Action](https://raw.githubusercontent.com/RonanOD/WordPresser/main/img/screenshot.jpeg)

## Features:
 * Go Routines for parallel lookup
 * OAuth2
 * File IO
 * HTTP Calls
 * JSON and REST handling
   * [WordPress API Docs](https://developer.wordpress.com/docs/api/)
 * Displays graphs and has interactive UI via [termui](https://github.com/gizak/termui)


## Needs:
 * Oauth2 Authentication set up. Follow details [here](https://developer.wordpress.com/docs/oauth2/)

