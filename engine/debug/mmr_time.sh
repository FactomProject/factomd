
(cat FNode0_missing_messages.txt; grep  "MissingMsg " FNode0_NetworkOutputs.txt | grep "Send P2P F"; grep -h "MissingMsgResponse" FNode*_NetworkOutputs.txt | grep "FNode0 ") | awk -f mmr_time.awk 

