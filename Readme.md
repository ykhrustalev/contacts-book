# How to use

* golang 1.11+


```
go run cmd/app/main.go list
go run cmd/app/main.go add -first-name John -last-name Smith -phone 11-00-01,11-00-02
go run cmd/app/main.go add -first-name Melissa -last-name "von Clark" -phone 22-00-01
go run cmd/app/main.go list
go run cmd/app/main.go edit -id 2 -first-name Melissa -last-name "von Clark" -phone 55-00-01
go run cmd/app/main.go list
go run cmd/app/main.go delete -id 2
go run cmd/app/main.go list
```
