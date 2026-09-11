# Mock

Mocking tool for REST APIs in Go.

## Install

```
go install github.com/dhuan/mock@latest
```

## Mock it

terminal 1
```
mock serve --port 3000 \
  --route 'say_hi/{name}' \
  --method GET \
  --response 'Hello, world! My name is ${name}.' \
  --route "what_time_is_it" \
  --method GET \
  --exec 'printf "Now it is %s" $(date +"%H:%M") | mock write'
```

terminal 2:
```
curl -i localhost:3000/say_hi/john_doe
curl -i localhost:3000/what_time_is_it 
```