# Moduł Tree – Instrukcja Użycia

Moduł `tree` służy do generowania struktury katalogów i plików w różnych formatach wyjściowych. Po ostatnich aktualizacjach narzędzie automatycznie agreguje reguły ignorowania plików z kilku poziomów oraz pozwala na limitowanie głębokości skanowania.

---

## Automatyczne Ignorowanie Plików (`.cmdrignore`)

Narzędzie automatycznie zbiera i unifikuje wzorce wykluczeń z następujących lokalizacji (w podanej kolejności):

1. **Globalna konfiguracja:** Plik `.cmdrignore` znajdujący się w Twoim katalogu profilu użytkownika (np. `~/.cmdr/.cmdrignore`).
2. **Katalog aplikacji:** Plik `.cmdrignore` umieszczony bezpośrednio obok pliku wykonywalnego `cmdr.exe`.
3. **Katalog roboczy (Lokalny):** Plik `.cmdrignore` w folderze, w którym aktualnie uruchamiasz komendę.

> **Wskazówka:** Reguły są automatycznie oczyszczane z końcowych ukośników (np. wpis `node_modules/` zadziała identycznie jak `node_modules`), dzięki czemu dopasowanie jest odporne na błędy w formatowaniu pliku ignore.

---

## Flagi i Opcje Komendy

```bash
cmdr tree [path] [flags]

```

### Dostępne flagi:

- `-d, --depth <int>` – Maksymalna głębokość skanowania drzewa:
- `-1` (Domyślnie) – Pełne drzewo, brak limitu głębokości.
- `0` – Tylko pliki i foldery w katalogu głównym (bieżący poziom).
- `1` – Katalog główny + jeden poziom podfolderów.

- `-f, --format <string>` – Format wyjściowy. Dostępne opcje: `ascii` (domyślny), `json`, `csv`, `md`.
- `-x, --exclude <strings>` – Dodatkowe, przekazywane w locie wzorce do wykluczenia (rozdzielane przecinkami).
- `-o, --output <file>` – Zapisuje wygenerowane drzewo bezpośrednio do pliku.
- `-c, --copy` – Kopiuje wygenerowaną strukturę bezpośrednio do schowka systemowego.
- `-g, --generate-ignore` – Generuje domyślny, globalny plik `.cmdrignore` w katalogu konfiguracyjnym aplikacji.

---

## Przykłady Użycia

### 1. Standardowe wyświetlenie struktury (pełne drzewo)

Wyświetli strukturę od bieżącego katalogu w dół, pomijając pliki zdefiniowane we wszystkich plikach `.cmdrignore`:

```bash
cmdr tree

```

### 2. Ograniczenie głębokości do konkretnego poziomu

Pokaże strukturę projektu w głąb tylko do 2 poziomów podkatalogów:

```bash
cmdr tree . -d 2

```

### 3. Zmiana formatu wyjściowego na Markdown i kopia do schowka

Generuje listę w formacie Markdown dla poziomu zerowego (tylko bieżący folder) i od razu wrzuca ją do schowka:

```bash
cmdr tree -f md -d 0 -c

```

### 4. Wykluczenie dodatkowych folderów "w locie" i zapis do pliku

Generuje strukturę w formacie JSON, dodatkowo ignorując foldery `internal` oraz `tests`, a wynik zapisuje w pliku `struktura.json`:

```bash
cmdr tree ./src -f json -x "internal,tests" -o struktura.json

```
