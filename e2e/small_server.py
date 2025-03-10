import http.server
import socket

class RequestHandler(http.server.BaseHTTPRequestHandler):
    def do_GET(self):
        # Get the hostname
        hostname = socket.gethostname()
        
        # Send response headers
        self.send_response(200)
        self.send_header('Content-type', 'text/plain')
        self.end_headers()
        
        # Write the hostname as the response
        self.wfile.write(f"Hostname: {hostname}.local".encode())

# Set up the server
server_address = ('', 8080)  # Listen on all available interfaces, port 8000
httpd = http.server.HTTPServer(server_address, RequestHandler)

print("Serving at http://localhost:8080/")
httpd.serve_forever()