## Still needs to be done

List of things that I still need to do in order to make GoESL completed
```
[x] Better documentation
[x] FreeSWITCH WIKI Golang page (proposal)
[ ] Unit testing
[ ] Add log/slog as default logger
[ ] Add reconnect logic
[ ] Add body option to SendEvent
    - Note:
        ```go
        	eventStrBuilder.WriteString("sendevent ")
	        eventStrBuilder.WriteString(*eventName)
	        eventStrBuilder.WriteString("\n")
	        eventStrBuilder.WriteString(*message)
        ```
[ ] More examples
```
