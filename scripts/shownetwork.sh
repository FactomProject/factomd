#!/bin/bash         
echo "; Default node color: Gray";
echo "{color:#555555}";
head -1000 out.txt | gawk -f scripts/showconnections.awk
echo "; Federated Servers: Blue";
echo "{color:#3333ff}"; 
tac out.txt | awk '/SummaryEnd/ {p=1; split($0, a, "SummaryEnd"); $0=a[1]}; /SummaryStart/ {p=0; split($0, a, "SummaryStart"); $0=a[2]; print; exit}; p' | tac | awk '{if ($2 ~ /FNode.*\[/) printf "%s\n", $0; else if ($3 ~ /FNode.*\[/) printf "%s\n", $0;}' | sed 's/\[/ /g' | awk '{if ($5 == "L") print $2; else if ($6 == "L") print $3;}' | sed 's/FNode//g'
echo "; Audit Servers: Green";
echo "{color:#338833}";
tac out.txt | awk '/SummaryEnd/ {p=1; split($0, a, "SummaryEnd"); $0=a[1]}; /SummaryStart/ {p=0; split($0, a, "SummaryStart"); $0=a[2]; print; exit}; p' | tac | awk '{if ($2 ~ /FNode.*\[/) printf "%s\n", $0; else if ($3 ~ /FNode.*\[/) printf "%s\n", $0;}' | sed 's/\[/ /g' | awk '{if ($4 == "A") print $2; else if ($5 == "A") print $3;}' | sed 's/FNode//g'
