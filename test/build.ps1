Set-Location ./auth
go build -o ./main.exe
Set-Location ../api
go build -o ./main.exe
Set-Location ../client
go build -o ./main.exe
Set-Location ../
