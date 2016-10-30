/^AAADMIN/ { 
    if ($6 > 5) off = 2
}

#/^dddd/ { not = 1 }

/^2/ { not = 1 }

#{ print off, $6, not, $0 }

!off && !not && $1 != "" {
   print
}
$1 == "" {
   off --
   if (off < 0) {
      off = 0
   }
}
{ not = 0 }
