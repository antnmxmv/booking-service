# Booking system

## About

This is a booking service prototype I have made for my job application to a Russian outstaff company.

I've decided to enhance it with more detailed documentation, usage examples and tests as it may help other developers somehow and surely it will help employers with getting what I am actually going to do as developer if they will hire me :)

## Documentation

Feel free to read the original task and my considerations on system and application design at [docs folder](docs/README.me).

## TODO
- [ ] switch to some helpful libraries where they needed
- [ ] use [watermill](https://github.com/ThreeDotsLabs/watermill) and any good ORM
- [ ] expose all ```BookingService``` methods in API
- [ ] at least 70% test coverage
- [ ] make it "kubernetes ready"

## Original task

HR manager provided me zip archive with task solutions which included:
- ```README.md``` file with wide task conditions
- ```main.go``` file with some data structures, helper functions and http request handler

### Restrictions
- usage of any libraries other than [chi router](https://github.com/go-chi/chi)
- solution must be easily switched to persistence

### P.S.

Note, that this is just my understanding of the problem. I've attached all original files with task conditions translated to English in this repo. [It's here.](docs/original-task/)
