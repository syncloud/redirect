import ssl
import subprocess
import threading
import time
from http.server import BaseHTTPRequestHandler, HTTPServer
from os import environ
from os.path import join

import requests
from syncloudlib.integration.hosts import add_host_alias

import api

BACKEND_BODY = 'relay-backend-ok'
BACKEND_PORT = 18443
FRPC = environ.get('FRPC', '/usr/local/bin/frpc')


class Handler(BaseHTTPRequestHandler):
    def do_GET(self):
        self.send_response(200)
        self.send_header('Content-Type', 'text/plain')
        self.end_headers()
        self.wfile.write(BACKEND_BODY.encode())

    def log_message(self, *args):
        pass


def gen_cert(work_dir):
    cert = join(work_dir, 'backend.crt')
    key = join(work_dir, 'backend.key')
    subprocess.check_call([
        'openssl', 'req', '-x509', '-newkey', 'rsa:2048',
        '-keyout', key, '-out', cert, '-nodes', '-days', '1',
        '-subj', '/CN=relay-backend'])
    return cert, key


def start_backend(work_dir):
    cert, key = gen_cert(work_dir)
    httpd = HTTPServer(('127.0.0.1', BACKEND_PORT), Handler)
    context = ssl.SSLContext(ssl.PROTOCOL_TLS_SERVER)
    context.load_cert_chain(cert, key)
    httpd.socket = context.wrap_socket(httpd.socket, server_side=True)
    threading.Thread(target=httpd.serve_forever, daemon=True).start()
    return httpd


def write_frpc_config(path, server_addr, server_name, token, domain_name):
    config = (
        'serverAddr = "{addr}"\n'
        'serverPort = 443\n'
        'transport.tls.enable = true\n'
        'transport.tls.serverName = "{sni}"\n'
        'transport.tls.disableCustomTLSFirstByte = true\n'
        'metadatas.token = "{token}"\n'
        '\n'
        '[[proxies]]\n'
        'name = "relay-e2e"\n'
        'type = "https"\n'
        'customDomains = ["{domain}"]\n'
        'localIP = "127.0.0.1"\n'
        'localPort = {port}\n'
    ).format(addr=server_addr, sni=server_name, token=token, domain=domain_name, port=BACKEND_PORT)
    with open(path, 'w') as f:
        f.write(config)


def start_frpc(work_dir, server_addr, server_name, token, domain_name, tag):
    config_path = join(work_dir, 'frpc-{0}.toml'.format(tag))
    log_path = join(work_dir, 'frpc-{0}.log'.format(tag))
    write_frpc_config(config_path, server_addr, server_name, token, domain_name)
    log = open(log_path, 'w')
    process = subprocess.Popen([FRPC, '-c', config_path], stdout=log, stderr=subprocess.STDOUT)
    return process, log_path


def fetch(domain_name):
    return requests.get('https://{0}/'.format(domain_name), verify=False, timeout=5)


def test_relay_valid_token_tunnels_traffic(domain, device_host, artifact_dir):
    user_domain = 'relaye2e'
    domain_name = '{0}.{1}'.format(user_domain, domain)
    token = api.domain_acquire(domain, domain_name, 'relay_e2e@syncloud.test', 'pass123456')
    add_host_alias(user_domain, device_host, domain)

    start_backend(artifact_dir)
    process, log_path = start_frpc(artifact_dir, device_host, 'relay.{0}'.format(domain), token, domain_name, 'valid')
    try:
        body = None
        for _ in range(30):
            try:
                response = fetch(domain_name)
                if response.status_code == 200:
                    body = response.text
                    break
            except Exception:
                pass
            time.sleep(2)
        assert body == BACKEND_BODY, open(log_path).read()
    finally:
        process.terminate()


def test_relay_bad_token_rejected(domain, device_host, artifact_dir):
    user_domain = 'relayneg'
    domain_name = '{0}.{1}'.format(user_domain, domain)
    add_host_alias(user_domain, device_host, domain)

    bad_token = '00000000-0000-0000-0000-000000000000'
    process, log_path = start_frpc(artifact_dir, device_host, 'relay.{0}'.format(domain), bad_token, domain_name, 'bad')
    try:
        time.sleep(10)
        got_backend = False
        try:
            got_backend = fetch(domain_name).text == BACKEND_BODY
        except Exception:
            got_backend = False
        assert not got_backend, 'relay served traffic for a domain the token does not own'
    finally:
        process.terminate()
