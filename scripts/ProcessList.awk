/ht:.*pl:.*FNode.*#VMs/ {
    stuff[on++]=$0
    on=1
}
on { stuff[on++]=$0 }
/Audit VMs:/ {
    for (i=0; i<on; i++) {
        print stuff[i]
    }
    on=0
}