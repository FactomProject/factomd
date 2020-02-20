/===SimElectionsEnd===/ {
    for (i=0; i<on1; i++) {
        print out1[i]
    }
    c1=0
}


c1 { out1[on1++]=$0 }

/===SimElectionsStart===/ {
    on1=0
    c1=1
}

