
build:
	go build -o bin/web cmd/web/main.go

test_get_view:
	curl -v -i -X GET http://localhost:4000/snippet/view

httpie_get_view:
	http GET http://localhost:4000/snippet/view

httpie_get_miss:
	http GET http://localhost:4000/miss

httpie_get_create:
	http GET http://localhost:4000/snippet/create


httpie_post_create:
	http POST http://localhost:4000/snippet/create

test_post_create:
	curl -i -X POST http://localhost:4000/snippet/create


test_get_with_id:
	curl -i -X GET http://localhost:4000/snippet/view?id=1234
