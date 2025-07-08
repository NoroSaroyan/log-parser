```bash
go run cmd/main.go corporate_resources/
go run cmd/main.go corporate_resources/mesrestapi2.log
go run cmd/main.go corporate_resources/mesrestapi.log corporate_resources/mesrestapi2.log
go run cmd/main.go --mode=parser corporate_resources/
go run cmd/main.go --mode=extractor corporate_resources/
go run cmd/main.go --mode=analyzer corporate_resources/


curl "http://localhost:8080/api/v1/download?pcbanumber=H8444A11100S60305140"
curl -i "http://localhost:8080/api/v1/pcbanumbers"
