# stupid blog api
мне лень делать фронтенд

### Post

`"POST /api/post"`
```
? create post
request:
	json{
		"title": "text" (required,min=3,max=255)
		"content": "content" (required,min=10)
	}
responce:
	json{
		"status": 2xx,
		"post_id": 123
	} if error {
		"status": error code,
		"error": "error text"
	}
```

"PATCH api/post/{id}"
```
request:
	json{
		"title": "text" required,min=3,max=255
		"content": "content" required,min=10
	}
response:
	json{
		"status": 2xx,
		"post_id": 123
	} if error {
		"status": error code,
		"error": "error text"
	}
```

`"DELETE /api/post/{id}"`
```
response:
	json{
		"status": 2xx,
		"post_id": 123
	} if error {
		"status": error code,
		"error": "error text"
	}
```

`"GET /api/post/{id}"`
```
response:
	json{
		"status": 2xx,
		"post_id": {id},
		"title": "post title",
		"content": "post content, some text...",
		"author_id": 4,
		"username": user123
	} if error {
		"status": error code,
		"error": "error text"
	}
```

`"GET /api/posts"`
```
? queries:
	limit - posts parse limit, default=10
	offset - posts parse offset, default=0

example-request: /api/posts?limit=10&offset=0 or /api/posts
response:
	json{
		"status": 200,
		"data": [
			{
				"post_id": 11,
				"title": "post title",
				"content": "content",
				"author_id": 2,
				"username": "user2",
				"created_at": "2025-06-27T13:39:55Z"
			},
			{
				"post_id": 9,
				"title": "post title",
				"content": "content 123123123123text",
				"author_id": 1,
				"username": "user1",
				"created_at": "2025-06-27T13:39:08Z"
			},...
		]
	} if error {
		"status": error code,
		"error": "error text"
	}
```

### User
`"POST /api/user/signup"`
```
"username": minlen=5
"password": minlen=8, valid=[A-Z, a-z, 1-9]
request:
	json{
		"username":"abc12",
		"email":"some@gmail.com",
		"password": "AbcPassword1234"
	}
response:
	json{
		"status": 2xx,
		"user": {
			"user_id": 7
		}
	} if error {
		"status": error code,
		"error": "error text"
	}
```

`"POST /api/user/signin"`
```
request:
	json{
		"username":"abc12",
		"email":"some@gmail.com",
		"password": "AbcPassword1234"
	}
response:
	json{
		"status": 202,
		"auth_token": "eyJhbGciOiJIUzI1NiIsInR5cC..."
	} if error {
		"status": error code,
		"error": "error text"
	}
```

### Comment

`"POST /api/comment"`
```
? create comment
"content": minlen=10,maxlen=1024
request:
	json{
		"post_id": 123,
		"content": "text"
	}
response:
	json{
		"status": 2xx,
		"comment_id,omitempty",
		"post_id,omitempty",
		"author_id,omitempty"`
	} if error {
		"status": error code,
		"error": "error text"
	}
```

`"DELETE /api/comment/{id}"`
```
request:
	json{
		"post_id": 123
	}
response:
	json{
		"status": 2xx,
		"comment_id": 1223,
		"author_id":  4
	} if error {
		"status": error code,
		"error": "error text"
	}
```

`"PATCH /api/comment/{id}"`
```
? {id} - comment id
"content": min=10,max=1024
request:
	json{
		"post_id": 123,
		"content": "new text"
	}
response:
	json{
		"comment_id": {id}
		"author_id": 4
		"post_id": 123
	} if error {
		"status": error code,
		"error": "error text"
	}

```

`"GET /api/comments"`
```
? queries:
	limit - comment parse limit, default=10
	offset - comment parse offset, default=0
	post_id - post id

response:
	json{
		"status": 200,
		"data": [
			{
				"comment_id": 77,
				"content": "NEW text",
				"post_id": 4,
				"author_id": 5,
				"username": "user5",
				"created_at": "2025-06-30T13:28:56Z"
			},
			{
				"comment_id": 78,
				"content": "some text",
				"post_id": 4,
				"author_id": 4,
				"username": "user4",
				"created_at": "2025-06-30T13:28:56Z"
			},...
		]
	} if error {
		"status": error code,
		"error": "error text"
	}
```
