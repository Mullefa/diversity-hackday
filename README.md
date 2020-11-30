# Diversity Hackday

The goal was to see if heuristics can be used to observe how demographic data relating to Guardian journalists has changed 
over time. To generate the heuristics, a predictive model was used to identify traits from a journalist's profile 
image. Within the permitted time, it was not feasible to build a custom model; instead the [facial analysis](https://aws.amazon.com/rekognition/image-features/)
feature of AWS' Rekognition service was used, which is able to predict age and gender. It is worth noting two caveats of
this service: gender is limited to male and female; and there has been [reported biases](https://www.theverge.com/2019/1/25/18197137/amazon-rekognition-facial-recognition-bias-race-gender)
in the model. Also note that to eliminate the potential to offend individuals, the heuristic data generated should be
anonymized. Practically speaking for this project, this means aggregating over groups with size large enough to 
preserve anonymity e.g. female / non-female. Outside of the confines of hackday, more work should also be done to 
anonymize the data at rest.

## Steps

1. Start Postgres locally:
   ```
   docker-compose up
   ```
2. Use the [Guardian's Open Platform](https://open-platform.theguardian.com/) to get a list of published articles and
   the journalists that wrote them; and write this data to Postgres:
   ```
   export CAPI_API_KEY=<api key>
   go run cmd/journalists-and-articles/main.go
   ```
3. Get the profile images of the journalists from [The Guardian website](https://www.theguardian.com/) and write these
   to local file system:
   ```
   go run cmd/journalist-image-scraper/main.go
   ```
4. Use the profile images and AWS Rekognition to generate heuristics on the journalists:
   ```
   go run cmd/heuristics/main.go
   ```
