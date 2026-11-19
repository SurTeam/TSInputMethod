#include "tsime_tsf.h"

#include <windows.h>

namespace {
HINSTANCE g_module = nullptr;
}

BOOL APIENTRY DllMain(HINSTANCE module, DWORD reason, LPVOID) {
  if (reason == DLL_PROCESS_ATTACH) {
    g_module = module;
    DisableThreadLibraryCalls(module);
  }
  return TRUE;
}

STDAPI DllGetClassObject(REFCLSID rclsid, REFIID riid, LPVOID *ppv) {
  if (ppv == nullptr) {
    return E_POINTER;
  }
  if (!IsEqualCLSID(rclsid, CLSID_TsimeTextService)) {
    return CLASS_E_CLASSNOTAVAILABLE;
  }
  auto *factory = new TsimeClassFactory();
  const HRESULT hr = factory->QueryInterface(riid, ppv);
  factory->Release();
  return hr;
}

STDAPI DllCanUnloadNow() { return S_OK; }

STDAPI DllRegisterServer() {
  wchar_t path[MAX_PATH];
  if (GetModuleFileNameW(g_module, path, MAX_PATH) == 0) {
    return HRESULT_FROM_WIN32(GetLastError());
  }
  const wchar_t *clsid = L"Software\\Classes\\CLSID\\{8F3C2A11-4B6D-4E2A-9C11-55AA22334455}";
  HKEY key = nullptr;
  if (RegCreateKeyExW(HKEY_CURRENT_USER, clsid, 0, nullptr, 0, KEY_WRITE, nullptr, &key, nullptr) != ERROR_SUCCESS) {
    return E_FAIL;
  }
  RegSetValueExW(key, nullptr, 0, REG_SZ, reinterpret_cast<const BYTE *>(L"TSInputMethod TSF"),
                 sizeof(L"TSInputMethod TSF"));
  HKEY inproc = nullptr;
  if (RegCreateKeyExW(key, L"InprocServer32", 0, nullptr, 0, KEY_WRITE, nullptr, &inproc, nullptr) == ERROR_SUCCESS) {
    RegSetValueExW(inproc, nullptr, 0, REG_SZ, reinterpret_cast<const BYTE *>(path),
                   static_cast<DWORD>((wcslen(path) + 1) * sizeof(wchar_t)));
    RegSetValueExW(inproc, L"ThreadingModel", 0, REG_SZ, reinterpret_cast<const BYTE *>(L"Apartment"),
                   sizeof(L"Apartment"));
    RegCloseKey(inproc);
  }
  RegCloseKey(key);
  return S_OK;
}

STDAPI DllUnregisterServer() {
  RegDeleteTreeW(HKEY_CURRENT_USER, L"Software\\Classes\\CLSID\\{8F3C2A11-4B6D-4E2A-9C11-55AA22334455}");
  return S_OK;
}
