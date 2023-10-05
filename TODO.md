## Still needs to be done

List of things that I still need to do in order to make GoESL completed

- [x] Better documentation
- [x] FreeSWITCH WIKI Golang page (proposal)
- [ ] Unit testing (in progress)
- [ ] Add log/slog as default logger
- [ ] Add Context
- [ ] Add reconnect logic
- [x] Add body option to SendEvent
  - Note:
    ```go
    eventStrBuilder.WriteString("sendevent ")
    eventStrBuilder.WriteString(*eventName)
    eventStrBuilder.WriteString("\n")
    eventStrBuilder.WriteString(*message)
    ```
- [ ] Add Job-UUID in SendMsg to Message Parse() so it's available as a header
- [ ] More examples
