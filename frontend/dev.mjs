import { createServer } from 'node:http';
import { readFile } from 'node:fs/promises';
import { extname, resolve } from 'node:path';

const root = resolve('.');
const mime = { '.html':'text/html', '.js':'text/javascript', '.css':'text/css' };
const server = createServer(async (req, res) => {
  let file = req.url === '/' ? 'index.html' : req.url.slice(1);
  try {
    const data = await readFile(resolve(root, file));
    res.writeHead(200, { 'Content-Type': mime[extname(file)] || 'application/octet-stream' });
    res.end(data);
  } catch {
    res.writeHead(404); res.end('Not found');
  }
});
server.listen(5173, '127.0.0.1', () => console.log('Name It frontend: http://127.0.0.1:5173'));
