"""使用临时文件验证转换结果，不接触已安装的 Typora。"""
import base64
import json
from pathlib import Path
import subprocess
import tempfile
import unittest

SCRIPT = Path(__file__).resolve().parents[1] / 'scripts' / 'typora.sh'


class TyporaShellTests(unittest.TestCase):
    def setUp(self):
        self.temp = tempfile.TemporaryDirectory(prefix='keyrun-typora-test-')
        self.addCleanup(self.temp.cleanup)
        self.root = Path(self.temp.name)

    def process(self, mode, *args):
        return subprocess.run(
            ['/bin/bash', '-c', 'source "$1"; shift; process_data "$@"',
             'test', str(SCRIPT), mode, str(self.root), *map(str, args)],
            capture_output=True,
        )

    def test_unicode_payload_and_rsa_signature(self):
        mc = {'l': '设备甲', 'i': 'fingerprint', 'v': '1.0'}
        (self.root / 'machine').write_bytes(base64.b64encode(json.dumps(mc, ensure_ascii=False).encode()))
        (self.root / 'email').write_text('中文+"test@example.com')
        (self.root / 'date').write_text('09/25/2026')
        result = self.process('payload')
        self.assertEqual(result.returncode, 0, result.stderr)
        payload = base64.b64decode((self.root / 'payload').read_bytes())
        entity = json.loads(payload)
        self.assertEqual(entity['deviceId'], mc['l'])
        self.assertEqual(entity['email'], '中文+"test@example.com')
        decoded = json.loads(payload.decode('latin1'))
        expected = ''.join(str(decoded[k]) for k in sorted(decoded)).encode()
        self.assertEqual((self.root / 'message').read_bytes(), expected)
        self.keys()
        subprocess.run(['/usr/bin/openssl', 'dgst', '-sha256', '-sign', str(self.root / 'private.pem'),
                        '-out', str(self.root / 'signature'), str(self.root / 'message')], check=True, capture_output=True)
        checked = subprocess.run(['/usr/bin/openssl', 'dgst', '-sha256', '-verify', str(self.root / 'public.pem'),
                                  '-signature', str(self.root / 'signature'), str(self.root / 'message')], capture_output=True)
        self.assertEqual(checked.returncode, 0, checked.stderr)

    def keys(self):
        subprocess.run(['/usr/bin/openssl', 'genrsa', '-out', str(self.root / 'private.pem'), '2048'], check=True, capture_output=True)
        subprocess.run(['/usr/bin/openssl', 'rsa', '-in', str(self.root / 'private.pem'), '-pubout',
                        '-out', str(self.root / 'public.pem')], check=True, capture_output=True)

    def test_binary_replacements_preserve_size_and_nulls(self):
        self.keys()
        original = SCRIPT.read_text().split("my $original = <<'KEY';\n")[1].split('\nKEY')[0].encode()
        source = self.root / 'input with spaces'
        dest = self.root / 'output'
        raw = b'\x00\xff' + original + b'\x00https://store.typora.io\x00' + original + b'https://dian.typora.com.cn\x00tail'
        source.write_bytes(raw)
        result = self.process('patch', source, dest)
        self.assertEqual(result.returncode, 0, result.stderr)
        patched = dest.read_bytes()
        self.assertEqual(len(patched), len(raw))
        self.assertEqual(source.read_bytes(), raw)
        self.assertEqual(patched.count((self.root / 'public.pem').read_bytes().rstrip()), 2)
        self.assertEqual(patched.count(b'https://127.0.0.1'), 2)
        self.assertTrue(patched.startswith(b'\x00\xff'))
        self.assertTrue(patched.endswith(b'\x00tail'))

    def test_incompatible_binary_is_untouched(self):
        self.keys()
        source, dest = self.root / 'input', self.root / 'output'
        source.write_bytes(b'incompatible\x00binary')
        result = self.process('patch', source, dest)
        self.assertNotEqual(result.returncode, 0)
        self.assertEqual(source.read_bytes(), b'incompatible\x00binary')
        self.assertFalse(dest.exists())

    def test_invalid_machine_code(self):
        for code in (b'not base64!', base64.b64encode(b'{}'), base64.b64encode(b'[]')):
            (self.root / 'machine').write_bytes(code)
            result = self.process('payload')
            self.assertNotEqual(result.returncode, 0)
            self.assertFalse((self.root / 'payload').exists())


if __name__ == '__main__':
    unittest.main()
