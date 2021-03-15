# WordPresser

Utility to query the stats of all of your WordPress sites from the command line. Drops a cache file of your OAuth2 access token. To reauthenticate, delete the `.token` file.

## Demo
![WordPresser In Action](https://raw.githubusercontent.com/RonanOD/WordPresser/main/img/demo.gif)

## Features:
 * OAuth2
 * File IO
 * HTTP Calls
 * JSON and REST handling
   * [WordPress API Docs](https://developer.wordpress.com/docs/api/)


## Needs:
 * Oauth2 Authentication set up. Follow details [here](https://developer.wordpress.com/docs/oauth2/)

## TODO:
 * Add visits struct for [graphing](https://github.com/gizak/termui)