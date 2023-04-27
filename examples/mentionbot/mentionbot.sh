#!/bin/bash

: ${MENTION:=ChatGPT}
LAST_MENTION=

while :; do
    read MSG
    if [[ "$MSG" = *${MENTION}* ]]; then
        NOW=$( date +'%s' )
        if [[ -n "$LAST_MENTION" ]]; then
            echo "It has been $(( NOW - LAST_MENTION )) seconds since someone mentioned ${MENTION} last!"
        fi
        LAST_MENTION="$NOW"
    fi

    sleep 0.1
done
