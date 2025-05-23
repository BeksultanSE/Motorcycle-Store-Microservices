 # UNIT test
cd order-service $$ go test ./internal/usecase -v 

# INTEGRATION  test
cd order-sece $$ go test ./tests/integration -v