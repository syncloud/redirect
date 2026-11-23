import os
import socket
import struct

MAGIC_COOKIE = 0x2112A442
BINDING_REQUEST = 0x0001
BINDING_SUCCESS = 0x0101
XOR_MAPPED_ADDRESS = 0x0020


def binding(host, port, timeout=5, attempts=5):
    transaction = os.urandom(12)
    request = struct.pack('!HHI', BINDING_REQUEST, 0, MAGIC_COOKIE) + transaction
    sock = socket.socket(socket.AF_INET, socket.SOCK_DGRAM)
    sock.settimeout(timeout)
    try:
        for _ in range(attempts):
            sock.sendto(request, (host, port))
            try:
                data, _ = sock.recvfrom(1500)
                break
            except socket.timeout:
                continue
        else:
            raise AssertionError('no stun reply from {0}:{1}'.format(host, port))
        local_port = sock.getsockname()[1]
    finally:
        sock.close()
    message_type, length, cookie = struct.unpack('!HHI', data[:8])
    assert message_type == BINDING_SUCCESS, hex(message_type)
    assert cookie == MAGIC_COOKIE
    assert data[8:20] == transaction
    return mapped_address(data[20:20 + length]), local_port


def mapped_address(attributes):
    offset = 0
    while offset + 4 <= len(attributes):
        attribute_type, attribute_length = struct.unpack('!HH', attributes[offset:offset + 4])
        value = attributes[offset + 4:offset + 4 + attribute_length]
        if attribute_type == XOR_MAPPED_ADDRESS:
            family = value[1]
            port = struct.unpack('!H', value[2:4])[0] ^ (MAGIC_COOKIE >> 16)
            assert family == 0x01, family
            address = struct.unpack('!I', value[4:8])[0] ^ MAGIC_COOKIE
            return socket.inet_ntoa(struct.pack('!I', address)), port
        offset += 4 + attribute_length + (-attribute_length % 4)
    raise AssertionError('no XOR-MAPPED-ADDRESS in stun reply')
