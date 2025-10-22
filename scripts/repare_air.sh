go env -w GOTOOLCHAIN=local
rm -f tmp/main
go build -o ./tmp/main ./cmd/api
file ./tmp/main