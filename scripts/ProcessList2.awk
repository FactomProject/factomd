/===ProcessListEnd===/{
    for (i=0; i<on; i++) {
        print stuff[i]
    }
    on=0
}

/===FederatedServersStart===/ { $0 = "Federated Servers:" }
/===FederatedServersEnd===/ { $0 = ""}
/===AuditServersStart===/ { $0 = "Audit Servers:"}
/===AuditServersEnd===/ { $0 = ""}


{
	if ($2 == "P") { 
		if (last) { 
			on--
		}
		last = 1
	}else{
		last = 0
	}
	
	if ( $4 == "EOM-VM" ) {
		last = 0
	}
}

on { stuff[on++]=$0 }



/===ProcessListStart===/{
    on=1
}


