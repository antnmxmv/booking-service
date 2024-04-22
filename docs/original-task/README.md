# Application design

The code provides a prototype of a hotel room booking service,
which allows you to book a free hotel room.

The service will grow, for example:

- a booking confirmation email will be sent
- discounts, promotional codes, loyalty programs will appear
- it will be possible to book multiple rooms

## Task

Refactor the structure and code of the application, fix existing logical
problems. There is no need to implement persistent storage. 
Store all data in the service memory.

As a result of the task, a structured service code is expected,
with correctly working logic for hotel room booking scenarios.

Check list:

- code reorganized and layers separated
- abstractions and interfaces are separated
- technical and logical errors fixed

Restrictions:

- we expect an implementation that manages the state in the application memory,
  but which can be easily replaced with external storage
- if you have experience with Go: for the solution you need to use only
  standard Go library + router (for example chi)
- if you have no experience with Go: you can implement the solution on your own
  favorite technology stack

## What will happen at the meeting

At the meeting, we expect you to share screen and present your solution:
Tell us what problems the source code has and how they are solved in your solution.
We will ask some questions about your decisions of dividing responsibility between
components in one way or another, and what principles you were guided by.
Let's ask what will happen if the product manager decides to add some new feature - how will it work
with the structure you proposed. We can also talk about more technical things:
about values and pointers, multithreading, interfaces, channels.

## Example

```sh
go run main.go
```

```sh
curl --location --request POST 'localhost:8080/orders' \
--header 'Content-Type: application/json' \
--data-raw '{
    "hotel_id": "reddison",
    "room_id": "lux",
    "email": "guest@mail.ru",
    "from": "2024-01-02T00:00:00Z",
    "to": "2024-01-04T00:00:00Z"
}'
```
