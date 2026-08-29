# cmdr - Bootstrap Prepare Password

Narzędzie CLI w języku Go służące do automatycznego skanowania plików konfiguracyjnych JSON oraz bezpiecznego generowania haseł.

### Podstawowe polecenie

```bash
cmdr bootstrap prepare-password --file config.json

```

---

## Skróty (Aliases)

Możesz użyć krótszych wersji komendy:

- `prep-pass`
- `pp`

Przykład:

```bash
cmdr bootstrap pp -f config.json

```

---

## Dostępne flagi

| Flaga      | Skrót | Domyślnie | Wymagana | Opis                                                |
| ---------- | ----- | --------- | -------- | --------------------------------------------------- |
| `--file`   | `-f`  | -         | **Tak**  | Ścieżka do pliku JSON, który ma zostać przetworzony |
| `--length` | `-l`  | `16`      | Nie      | Długość generowanego hasła                          |

---

## Przykłady działania

### 1. Plik przed przetworzeniem (`config.json`)

```json
{
  "app_name": "MyApp",
  "database": {
    "host": "localhost",
    "port": 5432,
    "password": 0
  },
  "services": [
    {
      "name": "auth-service",
      "password": "change_me"
    }
  ]
}
```

### 2. Wywołanie komendy z własną długością hasła (20 znaków)

```bash
cmdr bootstrap prepare-password -f config.json -l 20

```

### 3. Plik po przetworzeniu (`config.json`)

```json
{
  "app_name": "MyApp",
  "database": {
    "host": "localhost",
    "port": 5432,
    "password": "xK9#mP2!vL8@qW5$zR1T"
  },
  "services": [
    {
      "name": "auth-service",
      "password": "aB3&nQ9*pZ4#wX1^yU7M"
    }
  ]
}
```
