import struct
import socket
import time
import uuid

class Packet(object):
    __slots__ = [
        'op_code', 'priority', 'method', 'resource', 'sequence',
        'total_sequences', 'request_id', 'payload_length', 'payload'
    ]

    def __init__(self):
        self.op_code = 0x2
        self.priority = 0x5
        self.method = 0x00
        self.resource = 0x0000
        self.sequence = 0
        self.total_sequences = 0
        self.request_id = ''
        self.payload_length = 0
        self.payload = ''

    @classmethod
    def from_header(cls, header):
        packet = cls()
        packet.op_code = struct.unpack('!B', header[0])[0] >> 4
        packet.priority = struct.unpack('!B', header[0])[0] & 0x0F
        packet.method = struct.unpack('!B', header[1])[0] & 0x0F
        packet.resource = struct.unpack('!H', header[2:4])[0]
        packet.sequence = struct.unpack('!H', header[4:6])[0]
        packet.total_sequences = struct.unpack('!H', header[6:8])[0]
        packet.request_id = header[8:24]
        packet.payload_length = struct.unpack('!L', header[24:28])[0]
        return packet

class Request(object):
    __slots__ = ['op_code', 'method', 'resource', 'payload', 'request_uuid']

    def __init__(self, method, resource, payload, op_code=0x2):
        self.op_code = op_code
        self.method = method
        self.resource = resource
        self.payload = payload
        self.request_uuid = uuid.uuid4()

    def get_chunks(self):
        payload_length = len(self.payload)
        total_sequences = (payload_length / 1200) + 1

        for sequence in xrange(total_sequences):
            chunk_length = min(1200, payload_length - sequence * 1200)
            chunk = self.payload[sequence*1200:sequence*1200+chunk_length]

            yield struct.pack('!8B',
                (self.op_code << 4) | 0x5, self.method, self.resource >> 8, self.resource,
                sequence >> 8, sequence & 0xFF, total_sequences >> 8, total_sequences
            ) + self.request_uuid.bytes + struct.pack('!I', chunk_length) + chunk

class Teaspoon(object):
    def __init__(self, host, port):
        self.socket = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        self.socket.connect((host, port,))

    def close(self):
        self.socket.close()

    def recv_packets(self):
        while True:
            header = ''
            while len(header) < 28:
                header += self.socket.recv(28 - len(header))

            packet = Packet.from_header(header)

            packet.payload = ''
            while len(packet.payload) < packet.payload_length:
                packet.payload += self.socket.recv(packet.payload_length - len(packet.payload))

            yield packet

    def send_request(self, request):
        for chunk in request.get_chunks():
            self.socket.sendall(chunk)

        packets = []

        for packet in self.recv_packets():
            if packet.request_id != request.request_uuid.bytes:
                continue

            packets.append(packet)

            if packet.sequence == packet.total_sequences - 1:
                return Request(packet.method, packet.resource, ''.join([x.payload for x in packets]), op_code=packets[0].op_code)