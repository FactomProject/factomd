/FNode0.*L.*ID/ {
    collect=1
    once=1
    once1=1
}

/ min  0/ {
    if (once==1) {
        min=1
        collect=1
        once=0
    }
}

{
    if ($1=="ht:" && collect && once1==0) {
        stuff[on++]=$0
        once1==1
    }
}


collect { stuff[on++]=$0 }


/VM State per Node/ {

    collect=0
}

{
    if (min && $1==9) {
        for (i=0; i<on; i++) {
            print stuff[i]
        }
        collect=0
        on=0
        min=0
    }
 }