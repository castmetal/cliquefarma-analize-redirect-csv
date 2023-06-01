## Analize URLs to Redirect according SEO rules in Golang

1- Put file products_with_special_chars.csv on the root folder

> Run: go run main.go

2- Send all analisys into output.csv file

### Status Column

In the status collumn the value: REDIRECIONAR, it will be the values that will have to REDIRECT, they are all URLs that will be change on the app.

ALTERAR: There are URLs that having 404 status today and they need to be changed on the DATABASE due to special characters.

ANALISAR: There are URLs that having status 200 ocorre into from URLs, in other words on it need to be change, because its wrong.
