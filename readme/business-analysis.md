# Time tracking webpage

## Background
Lots of small companies wish to have overview of time spent by their workforce.
It can be done via Jira etc. but these are usually costly solutions.
We try to give more simple and cheaper alternative to track time spend coding and possibly 
in other similar activities. 

## Components
### Jetbrais / etc. plugins
IDE plugins allow capturing events without extra steps.

### CLI app for client machine
CLI app for storing tracked data and syncing it to server.
CLI app should be easy enough to be used without any plugins to have brief overview.

### Backend
Backend should store data of each team member. Accessing data should be restricted to
only authorized persons. Backend should also provide api for external usage.

### Web app
Web app is for displaying data via fancy graphs and some numbers. 
It should be possible to login as single user as well as a company. Users can see
all their time tracked. This means if a person is working for multiple companies et, he can
still have overview of data in single place. Companies have only time spent at work to 
avoid evading private lives.
#### Landig page
TODO
#### Login page
TODO
#### TODO...
 
 ## User stories
  - [ ] As a programmer I can have my time tracked without extra actions
  - [ ] As a programmer I can view my time spend and compare it to others / average
  - [ ] As a team lead I can have detailed view of my team members time spent on different activities
  - [ ] As a team lead I can view time spent on each issue
  - [ ] As a tester I can have overview of how many times I have run tests / app
 