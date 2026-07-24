import ssl
import subprocess
import tempfile
import threading
import time
from http.server import BaseHTTPRequestHandler, HTTPServer
from os import environ
from os.path import join

import requests
from syncloudlib.integration.hosts import add_host_alias

import api
from test import create_user, get_domain

RELAY_ADDRESS = '10.0.0.99'

BACKEND_BODY = 'relay-backend-ok'
BACKEND_PORT = 18443
BIG_BODY = ('x' * 65536).encode()
FRPC = environ.get('FRPC', '/usr/local/bin/frpc')


class Handler(BaseHTTPRequestHandler):
    def do_GET(self):
        body = BIG_BODY if self.path == '/big' else BACKEND_BODY.encode()
        self.send_response(200)
        self.send_header('Content-Type', 'text/plain')
        self.send_header('Content-Length', str(len(body)))
        self.end_headers()
        self.wfile.write(body)

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
        'name = "{domain}"\n'
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
    email = 'relay_e2e@syncloud.test'
    password = 'pass123456'
    create_user(domain, email, password, artifact_dir)
    token = api.domain_acquire(domain, domain_name, email, password)
    add_host_alias(user_domain, device_host, domain)

    work_dir = tempfile.mkdtemp()
    backend = start_backend(work_dir)
    process, log_path = start_frpc(work_dir, device_host, 'relay.{0}'.format(domain), token, domain_name, 'valid')
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
        backend.shutdown()


def test_relay_bad_token_rejected(domain, device_host):
    user_domain = 'relayneg'
    domain_name = '{0}.{1}'.format(user_domain, domain)
    add_host_alias(user_domain, device_host, domain)

    work_dir = tempfile.mkdtemp()
    bad_token = '00000000-0000-0000-0000-000000000000'
    process, log_path = start_frpc(work_dir, device_host, 'relay.{0}'.format(domain), bad_token, domain_name, 'bad')
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


def test_relay_update_points_dns_at_relay(domain, artifact_dir):
    email = 'relay_dns@syncloud.test'
    password = 'pass123456'
    create_user(domain, email, password, artifact_dir)
    domain_name = 'relaydns.{0}'.format(domain)
    update_token = api.domain_acquire(domain, domain_name, email, password)

    response = requests.post('https://api.{0}/domain/update'.format(domain), json={
        'token': update_token,
        'ipv4_enabled': True,
        'relay': True,
        'web_protocol': 'https',
        'web_local_port': 443,
    }, verify=False)
    assert response.status_code == 200, response.text

    data = get_domain(update_token, domain)
    assert data['ip'] == RELAY_ADDRESS, data


def test_relay_monthly_limit_blocks_traffic(domain, device_host, artifact_dir):
    user_domain = 'relayquota'
    domain_name = '{0}.{1}'.format(user_domain, domain)
    email = 'relay_quota@syncloud.test'
    password = 'pass123456'
    create_user(domain, email, password, artifact_dir)
    token = api.domain_acquire(domain, domain_name, email, password)
    add_host_alias(user_domain, device_host, domain)

    work_dir = tempfile.mkdtemp()
    backend = start_backend(work_dir)
    process, log_path = start_frpc(work_dir, device_host, 'relay.{0}'.format(domain), token, domain_name, 'quota')
    try:
        up = False
        for _ in range(30):
            try:
                if fetch(domain_name).status_code == 200:
                    up = True
                    break
            except Exception:
                pass
            time.sleep(2)
        assert up, open(log_path).read()

        time.sleep(4)

        for _ in range(5):
            try:
                requests.get('https://{0}/big'.format(domain_name), verify=False, timeout=5)
            except Exception:
                pass

        blocked = False
        for _ in range(20):
            try:
                if requests.get('https://{0}/big'.format(domain_name), verify=False, timeout=5).status_code != 200:
                    blocked = True
                    break
            except Exception:
                blocked = True
                break
            time.sleep(1)
        assert blocked, 'relay kept serving after exceeding the monthly limit\n' + open(log_path).read()
    finally:
        process.terminate()
        backend.shutdown()
