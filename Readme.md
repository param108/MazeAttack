# Maze Attack

A simple server that controls a player on a maze.

## Setup
Clone into gopath. Then do

`go build main.go`

## Running

./main

## APIs

Create a user with name "xxxxx". Users once created are never deleted!
```
POST to `/create/` with
{
"user":"xxxxx"
}

Response
{
"Error":<string>,
"Message":<string>,
"X":<int>
"Y":<int>
"BaddieX":<int>
"BaddieY":<int>
}

X,Y are the players current position
BaddieX, BaddieY are the destination position
```

Move a user on the board with
```
POST to `/move/` with
{
"user":"xxxxx"
"move":"dir"
}

dir is one of UP | DOWN | RIGHT | LEFT

Response
{
"Error":<string>,
"Message":<string>,
"X":<int>
"Y":<int>
"BaddieX":<int>
"BaddieY":<int>
}

X,Y are the players position after the move
BaddieX, BaddieY are the destination position
```

## Notes
1. If you hit a wall you will just remain where you are no. No special response is given.
1. The only way to know if you have succeeded is if (BaddieX == X && BaddieY == Y). The Server will print the message "xxxx has won".

## Sample Maze

`# are walls and E is the destination (Baddie)`
```
   #   # #       #     # #       # # # #   # #
         # #     #     #       # # #     #   #   #
 #   #   # #   # # # #         # # # #       # #
   # #   # # # # #   # #   # #   #   # # #     # #
 # # # #   # #   #   # #   #   #   #   #   #   #
 #       #   # #   #   # #   #   # # #     # # #
         #   # # #   #     #     # # # #   #   # #
 # #     #       # # # #   #     # # #       #   #
 # #   # #       # #   # #   # # #         # # # #
 # # #     #           # #     #   #   # #       #
   #   #     #   #   # #     #               #
 # #         # # #       # #   # #   #   #   #
     # #     #   # # #       # #       #
   #         # #     #   # # #     # # # #     # #
 #   # #   # #   #   # #   #   # #     #   #   #
 #         #   # #         #       # #       #   #
 # #       #   # #     # #       #   # # #   #
 # # # #       # #   #       # # # # #     #
 # # # # # #   # #   # # #   # # #   #   #   #
 # # #   # #   #   # # # #   # #     #   #   #   #
     #   #       #   #   # #   #   #   #   # #   #
     # # # #     # #           # #   #     # #   #
   #     #         #   #     #     #         #
     # #   #       #   # # # # # #   #   #   #   #
 # #   #   # #           #       # #   #   # #   #
             #   # #     # # # #   # # #   # #   #
     # # # #   #   #     #   #     #     #
 #     # #   #   #   # # # #   # #         #     #
     #         #   # #       #   # # # # #   # #
 #   #       # # # # # # #             # # #
     #   #     # # # # # #   # # # # #       # # #
   # # #   #   #     #     #     #   # #     # # #
   # # # #       #   # #   # # #     #   # #   # #
     # # # # #   #             #     # # #   #   #E
 #   # #     # # # #   # #   # # # #     # # # # #
   #   #       # #   #     #       #           #
 #   #       # #   #   # #             # # #
 #   # #   #   # #   #   # # # #   # # # # #   # #
 #       # #   # #     # #   # #   # # # #     #
 # #   #         #               # # #     #   # #
           #       # #   #   # #   #           #
   #     # # #           # #     # #     #     #
   #       #   #   # #   #     # #   # # #     # #
   #     #   # #           # #     #
   #   #           #   #   #   # #   # #       # #
 #     #   #   #       #   # #   #   # # # #   # #
 # #   # # #                   # #   #         # #
 #         #   #         #       # #     # # #
   # # # # #   #       #       # # # #     # #   #
   # #   # #       # #   #         # # # #   # #
```
