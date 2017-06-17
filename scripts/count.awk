/Buying/ { fcts++ }
/create chain/ { chains++ }
/Entryhash:/ { entries++ }
END {
  print "Fct trans  ", fcts
  print "Chains created  ", chains
  print "Entries created ", entries
  print ""
}
