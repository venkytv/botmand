#!/bin/bash

NGUESSES=0
guess_the_number() {
    NUMBER=$(( ( RANDOM % 20 )  + 1 ))

    echo "I've picked a number between 1 and 20. What is it?"

    while true; do
        echo "Your guess?"
        while true; do
            read GUESS
            if [[ $GUESS =~ ^[0-9]+$ ]]; then
                break
            else
                echo "Please enter a number."
            fi
        done
        NGUESSES=$(( NGUESSES + 1 ))

        if [[ $GUESS -eq $NUMBER ]]; then
            echo "You got it! The number was $NUMBER."
            break
        elif [[ $GUESS -lt $NUMBER ]]; then
            echo "Too low!"
        else
            echo "Too high!"
        fi
    done
}

rock_paper_scissors() {
    OPTIONS=( "Rock" "Paper" "Scissors" )
    CHOICE=${OPTIONS[$(( RANDOM % 3 ))]}

    echo "Choose one:
    1) Rock
    2) Paper
    3) Scissors" | literal_newlines
    echo "Enter your choice [1-3]"

    while true; do
        read USER_CHOICE
        if [[ $USER_CHOICE -ge 1 && $USER_CHOICE -le 3 ]]; then
            break
        else
            echo "Please enter a number between 1 and 3."
        fi
    done
    USER_CHOICE=${OPTIONS[$(( USER_CHOICE - 1 ))]}

    echo "You chose $USER_CHOICE. I chose $CHOICE."
    if [[ $USER_CHOICE == $CHOICE ]]; then
        echo "It's a tie!"
    elif [[ $USER_CHOICE == "Rock" && $CHOICE == "Scissors" ]]; then
        echo "You win!"
    elif [[ $USER_CHOICE == "Paper" && $CHOICE == "Rock" ]]; then
        echo "You win!"
    elif [[ $USER_CHOICE == "Scissors" && $CHOICE == "Paper" ]]; then
        echo "You win!"
    else
        echo "I win!"
    fi
}

literal_newlines() {
    # If output is a terminal, don't replace newlines
    # Otherwise, replace newlines with literal "\n"
    if [[ -t 1 ]]; then
        awk '{printf "%s\n", $0}'
    else
        awk '{printf "%s\\n", $0}'
    fi
}

echo "Welcome to GameBot!
Which game would you like to play?
1) Guess the number
2) Rock, Paper, Scissors
Enter your choice [1-2]:" | literal_newlines

while true; do
    read GAME_CHOICE
    if [[ $GAME_CHOICE -ge 1 && $GAME_CHOICE -le 2 ]]; then
        break
    else
        echo "Please enter a number between 1 and 2."
    fi
done

if [[ $GAME_CHOICE -eq 1 ]]; then
    echo "*Guess the Number!* botmand://switch/thread"  # Start the game in a new thread
    NGUESSES=0
    guess_the_number

    # Post results to the channel
    echo "Announcing results in the channel botmand://switch/channel"
    echo "The number was guessed in $NGUESSES attempts."
else
    echo "*Rock, Paper, Scissors!* botmand://switch/thread" # Start the game in a new thread
    rock_paper_scissors
fi
