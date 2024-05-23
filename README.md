# goilerplate
A boilerplate for Golang Restful API with auto-reload for every change that happens on code.



## Attention please

### Project structure
| main\
    | config\
    | api
        | middleware
        | request
        | response
        | route 
        | ...
    | app
        | entity
        | handler
            | v1
            | ...
        | usecase 
            | v1
            | ...
        | repository
            | v1
            | ...
        | helper 
        | logs

##

Alyaws remember the rules: **Avoid importing packages at the same or higher layer to avoid circular import problems**

### Project hierarchy

| main
    | api
        | route
            | handler 
                | usecase 
                    | repository
                        | middleware
                        | request
                        | response 
                            | config
                                | entity
                                    | helper

## 

                                    




        

