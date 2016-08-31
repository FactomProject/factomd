/===SummaryEnd===/ {
    for (i=0; i<on2; i++) {
        print out2[i]
    }

    for (i=0; i<on1; i++) {
        print out1[i]
    }
    c1=0
}

/===PrintMapEnd===/ {
    c2=0
}

c1 { out1[on1++]=$0 }
c2 { out2[on2++]=$0 }

/===SummaryStart===/ {
    on1=0
    c1=1
}
/===PrintMapStart===/ {
    on2=0
    c2=1
    out2[on2++]=$0
}

