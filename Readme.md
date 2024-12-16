## Start application
### flags 
- --workers "для задания числа горутин."
- --input "для указания пути к файлу с URL"
- --timeout "для настройки таймаута HTTP-запросов." (по умолчанию 5)

* go run main.go --workers=3 --input=urls.txt --timeout=10