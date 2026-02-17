#!/usr/bin/env bash
# check_config.sh – Verifikuje da sve putanje u config.yaml postoje na disku
# Pokreni: bash check_config.sh [path/to/config.yaml]

set -euo pipefail

CONFIG="${1:-config.yaml}"
RED='\033[0;31m'; GREEN='\033[0;32m'; YELLOW='\033[1;33m'; NC='\033[0m'

echo -e "${YELLOW}Provera config fajla: $CONFIG${NC}"
echo ""

if [ ! -f "$CONFIG" ]; then
    echo -e "${RED}✗ Config fajl ne postoji: $CONFIG${NC}"
    exit 1
fi

# Parsiramo YAML ručno (bez dodatnih alata)
current_profile=""
errors=0

while IFS= read -r line; do
    # Detektuj profil (red koji ima 2 razmaka i završava sa ":")
    if echo "$line" | grep -qE '^  [a-zA-Z0-9_]+:$'; then
        current_profile=$(echo "$line" | tr -d ' :')
        echo -e "  ${YELLOW}Profil: $current_profile${NC}"
    fi

    # Proveri peerTLSCert
    if echo "$line" | grep -qE '^\s+peerTLSCert:'; then
        path=$(echo "$line" | sed 's/.*peerTLSCert: *//' | tr -d '"')
        if [ -z "$path" ]; then
            echo -e "    ${RED}✗ peerTLSCert je prazan!${NC}"
            ((errors++))
        elif [ -f "$path" ]; then
            echo -e "    ${GREEN}✓ peerTLSCert:${NC} $path"
        else
            echo -e "    ${RED}✗ peerTLSCert ne postoji:${NC} $path"
            ((errors++))
        fi
    fi

    # Proveri certPath
    if echo "$line" | grep -qE '^\s+certPath:'; then
        path=$(echo "$line" | sed 's/.*certPath: *//' | tr -d '"')
        if [ -z "$path" ]; then
            echo -e "    ${RED}✗ certPath je prazan!${NC}"
            ((errors++))
        elif [ -f "$path" ]; then
            echo -e "    ${GREEN}✓ certPath:${NC} $path"
        else
            echo -e "    ${RED}✗ certPath ne postoji:${NC} $path"
            ((errors++))
        fi
    fi

    # Proveri keyPath (može biti fajl ili direktorijum)
    if echo "$line" | grep -qE '^\s+keyPath:'; then
        path=$(echo "$line" | sed 's/.*keyPath: *//' | tr -d '"')
        if [ -z "$path" ]; then
            echo -e "    ${RED}✗ keyPath je prazan!${NC}"
            ((errors++))
        elif [ -e "$path" ]; then
            # Ako je direktorijum, provjeri da ima bar jedan fajl
            if [ -d "$path" ]; then
                count=$(ls "$path" 2>/dev/null | wc -l)
                if [ "$count" -gt 0 ]; then
                    echo -e "    ${GREEN}✓ keyPath (dir, $count fajl/ova):${NC} $path"
                else
                    echo -e "    ${RED}✗ keyPath direktorijum je prazan:${NC} $path"
                    ((errors++))
                fi
            else
                echo -e "    ${GREEN}✓ keyPath (fajl):${NC} $path"
            fi
        else
            echo -e "    ${RED}✗ keyPath ne postoji:${NC} $path"
            ((errors++))
        fi
    fi

done < "$CONFIG"

echo ""
if [ "$errors" -eq 0 ]; then
    echo -e "${GREEN}══ Sve putanje su OK! ══${NC}"
else
    echo -e "${RED}══ Pronađeno $errors grešaka! ══${NC}"
    echo ""
    echo "Saveti:"
    echo "  1. Proveri da li je mreža pokrenuta: cd fabric-samples/test-network && ./network.sh up"
    echo "  2. Koristi 'find' da pronađeš tačno ime fajla:"
    echo "     find .../organizations -name 'cert.pem' | grep Admin"
    echo "     find .../organizations -name 'ca.crt' | grep peer0"
    exit 1
fi