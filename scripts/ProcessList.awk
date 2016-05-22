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


on { stuff[on++]=$0 }

/===ProcessListStart===/{
    on=1
}


