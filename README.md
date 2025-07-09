```bash
go run cmd/main.go corporate_resources/
go run cmd/main.go corporate_resources/mesrestapi2.log
go run cmd/main.go corporate_resources/mesrestapi.log corporate_resources/mesrestapi2.log
go run cmd/main.go --mode=parser corporate_resources/
go run cmd/main.go --mode=extractor corporate_resources/
go run cmd/main.go --mode=analyzer corporate_resources/
```
```bash
# Get list of PCBA numbers
curl -i "http://localhost:8080/api/v1/pcbanumbers"

# Download info for a specific PCBA number
curl -i "http://localhost:8080/api/v1/download?pcbanumber=H8444A11100S60305140"

# Final info for a specific PCBA number
curl -i "http://localhost:8080/api/v1/final?pcbanumber=H8444A11100S60305140"

# PCBA info for a specific PCBA number
curl -i "http://localhost:8080/api/v1/pcba?pcbanumber=H8444A11100S60305140"
```