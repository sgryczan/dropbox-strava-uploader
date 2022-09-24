# DropBox Strava Uploader

## What is this?
This is a little daemon that reads .gpx files from Dropbox and uploads them to strava

## Why?
I ride my e-bike around a LOT and like to record my progress in [wandrer.earth](https://wandrer.earth). Wandrer imports rides automatically from strava, but since I don't record my rides using the strava app (I record the .gpx files myself using a different application), importing the .gpx files from my phone to the computer and uploading them to strava is a tedious process. Each ride I record is automatically uploaded to Dropbox, and this this program will read the files from there and upload them to strava.

## Using the uploader
1. Get a [token for your dropbox account](https://dropbox.tech/developers/generate-an-access-token-for-your-own-account).
2. Get a [client id and secret for your strava account](https://developers.strava.com/docs/getting-started/#account).
    * Set the `Authorization callback domain` to whatever domain resolves to where the image is running. If you're running this locally, use `localhost`
3. Build the docker image by running `make`
4. Run the image:
```
$ docker run \
    -e "STRAVA_CLIENT_ID=<your strava client id>" \
    -e "STRAVA_CLIENT_SECRET=<your strava client secret>" \
    -e "CALLBACK_DOMAIN=<your callback domain>" \
    -e "DROPBOX_TOKEN=<your dropbox token>" \
    -p 8321:8321 \
    -v $pwd:/app \
    sgryczan/strava-uploader:latest
```
5. Navigate to localhost:8321 in a browser, and click on the top-left corner.
    * You'll be taken to a strava authorization page. Click 'authorize'.
