```bash
go run cmd/main.go corporate_resources/
go run cmd/main.go corporate_resources/mesrestapi2.log
go run cmd/main.go corporate_resources/mesrestapi.log corporate_resources/mesrestapi2.log
go run cmd/main.go --mode=parser corporate_resources/
go run cmd/main.go --mode=extractor corporate_resources/
go run cmd/main.go --mode=analyzer corporate_resources/
