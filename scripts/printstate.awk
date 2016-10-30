/===PrintState End===/{
    for (i=0; i<on; i++) {
        print stuff[i]
    }
    on=0
}


on { stuff[on++]=$0 }

/===PrintState Start===/{
    on=1
}


