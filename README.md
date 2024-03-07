# offline-browsing

## Run Instructions

1. Build the image: ```docker build -t fetch .```
2. Run the container: ```docker run --name fetch -t -d fetch```
3. Execute the command: ```docker exec -it fetch sh -c "./fetch -h"```
