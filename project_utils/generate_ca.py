import os
import subprocess

def generate_ca(base_dir):
    ca_dir = os.path.join(base_dir, "certificates", "CA")
    
    # Create directories if they don't exist
    os.makedirs(ca_dir, exist_ok=True)

    # Generate CA key
    subprocess.run([
        "openssl", "genpkey", "-algorithm", "RSA", "-out", os.path.join(ca_dir, "ca.key")
    ], check=True)

    # Generate CA certificate
    subprocess.run([
        "openssl", "req", "-new", "-x509", "-key", os.path.join(ca_dir, "ca.key"),
        "-out", os.path.join(ca_dir, "ca.pem"), "-days", "365", "-subj", "/CN=Test CA"
    ], check=True)

def main():
    # Generate CA for the main directory
    generate_ca(".")
    
    # Generate CA for the tests directory
    generate_ca("tests")

if __name__ == "__main__":
    main()