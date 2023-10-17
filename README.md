# leaderboard-api
A leaderboard API to store scores and efficiently query for rankings. Built with Go (+gorilla/mux) &amp; Redis 


## Key Takeaways:
- API Development (model, controller, routing)
- Documentation (OpenAPI standard, Swagger, code annotations)
- Redis (Docker containerization, go-redis client)
- JSON (marshalling)
- Testing ()

### For personal reference:
Make sure you have swaggo installed on your machine:

`$ go install github.com/swaggo/swag/cmd/swag@latest`

To generate swagger documentation automatically 

`$ swag init`
 
swagger.json and swagger.yaml should now reside under the docs folder.


### TODOs:
- [ ] testing framework for the API
- [ ] add endpoint for rangequery
- [ ] incorporate pagination, filtering properties
- [ ] make server run on goroutine, add timeout functionality
- [ ] stress test, optimize running times


