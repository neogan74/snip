test_get_view:
	curl -v -i -X GET http://localhost:4000/snippet/view

httpie_get_view:
	http GET http://localhost:4000/snippet/view

httpie_get_miss:
	http GET http://localhost:4000/miss

httpie_post_create:
	http POST http://localhost:4000/snippet/create