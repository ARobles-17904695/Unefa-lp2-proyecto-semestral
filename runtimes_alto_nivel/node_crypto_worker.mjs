import { Worker, isMainThread, parentPort, workerData } from 'node:worker_threads';
import { pbkdf2Sync } from 'node:crypto';
import { fileURLToPath } from 'node:url';

if (isMainThread) {
  console.log('[Main Thread] Iniciando cálculo criptográfico en Worker Thread...');

  // Temporizador para demostrar que el Event Loop NO se bloquea
  const heartbeat = setInterval(() => {
    console.log('[Main Thread] Event Loop libre y respondiendo...');
  }, 40);

  const worker = new Worker(fileURLToPath(import.meta.url), {
    workerData: {
      password: 'clave_super_secreta_transaccion',
      salt: 'sal_aleatoria_segura_2026',
      iterations: 350000,
      keylen: 64,
      digest: 'sha256',
    },
  });

  worker.on('message', (resultado) => {
    clearInterval(heartbeat);
    console.log(`[Main Thread] Hash PBKDF2 recibido (${resultado.tiempoMs} ms):`);
    console.log(resultado.hash.slice(0, 48) + '...');
  });

  worker.on('error', (err) => {
    clearInterval(heartbeat);
    console.error('Error en Worker:', err);
  });
} else {
  const inicio = performance.now();
  const { password, salt, iterations, keylen, digest } = workerData;
  const derivedKey = pbkdf2Sync(password, salt, iterations, keylen, digest);
  const tiempoMs = (performance.now() - inicio).toFixed(2);

  parentPort.postMessage({
    hash: derivedKey.toString('hex'),
    tiempoMs,
  });
}