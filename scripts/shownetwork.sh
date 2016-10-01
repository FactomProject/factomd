#!/bin/bash         
echo "; Default node color: Gray";
echo "{color:#555555}";
head -1000 out.txt | gawk -f scripts/showconnections.awk
echo "; Federated Servers: Blue";
echo "{color:#3333ff}"; 
tail -10000 out.txt | gawk -f scripts/status.awk | awk '{if ($2 ~ /FNode.*\[/) printf "%s\n", $0; else if ($3 ~ /FNode.*\[/) printf "%s\n", $0;}' | sed 's/\[/ /g' | awk '{if ($5 == "L") print $2; else if ($6 == "L") print $3;}' | sed 's/FNode//g'
echo "; Audit Servers: Green";
echo "{color:#338833}";
tail -10000 out.txt | gawk -f scripts/status.awk | awk '{if ($2 ~ /FNode.*\[/) printf "%s\n", $0; else if ($3 ~ /FNode.*\[/) printf "%s\n", $0;}' | sed 's/\[/ /g' | awk '{if ($4 == "A") print $2; else if ($5 == "A") print $3;}' | sed 's/FNode//g'
