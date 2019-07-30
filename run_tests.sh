gofmt -s -w .
go-acc ./... -o cover.out --covermode=set # -count for fancy
go tool cover -html=cover.out -o cover.html
rm cover.out
