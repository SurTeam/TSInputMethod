#include "tsime_tsf.h"

#include <string>

namespace {

std::wstring utf8ToWide(const std::string &value) {
  if (value.empty()) {
    return L"";
  }
  const int size = MultiByteToWideChar(CP_UTF8, 0, value.c_str(), -1, nullptr, 0);
  std::wstring out(size > 0 ? size - 1 : 0, L'\0');
  if (size > 1) {
    MultiByteToWideChar(CP_UTF8, 0, value.c_str(), -1, out.data(), size);
  }
  return out;
}

bool isAsciiLetterOrEdit(WPARAM wParam) {
  if (wParam == VK_BACK || wParam == VK_ESCAPE || wParam == VK_SPACE) {
    return true;
  }
  if (wParam >= '0' && wParam <= '9') {
    return true;
  }
  return (wParam >= 'A' && wParam <= 'Z') || (wParam >= 'a' && wParam <= 'z');
}

} // namespace

const CLSID CLSID_TsimeTextService = {
    0x8f3c2a11, 0x4b6d, 0x4e2a, {0x9c, 0x11, 0x55, 0xaa, 0x22, 0x33, 0x44, 0x55}};

TsimeTextService::TsimeTextService() = default;

TsimeTextService::~TsimeTextService() {
  Deactivate();
}

STDMETHODIMP TsimeTextService::QueryInterface(REFIID riid, void **ppvObj) {
  if (ppvObj == nullptr) {
    return E_INVALIDARG;
  }
  if (IsEqualIID(riid, IID_IUnknown) || IsEqualIID(riid, IID_ITfTextInputProcessor) ||
      IsEqualIID(riid, IID_ITfTextInputProcessorEx)) {
    *ppvObj = static_cast<ITfTextInputProcessorEx *>(this);
  } else if (IsEqualIID(riid, IID_ITfKeyEventSink)) {
    *ppvObj = static_cast<ITfKeyEventSink *>(this);
  } else {
    *ppvObj = nullptr;
    return E_NOINTERFACE;
  }
  AddRef();
  return S_OK;
}

STDMETHODIMP_(ULONG) TsimeTextService::AddRef() { return InterlockedIncrement(&ref_); }

STDMETHODIMP_(ULONG) TsimeTextService::Release() {
  const LONG value = InterlockedDecrement(&ref_);
  if (value == 0) {
    delete this;
  }
  return value;
}

STDMETHODIMP TsimeTextService::Activate(ITfThreadMgr *ptim, TfClientId tid) {
  return ActivateEx(ptim, tid, 0);
}

STDMETHODIMP TsimeTextService::ActivateEx(ITfThreadMgr *ptim, TfClientId tid, DWORD) {
  if (ptim == nullptr) {
    return E_INVALIDARG;
  }
  threadMgr_ = ptim;
  threadMgr_->AddRef();
  clientId_ = tid;
  ITfKeystrokeMgr *keyMgr = nullptr;
  if (SUCCEEDED(threadMgr_->QueryInterface(IID_ITfKeystrokeMgr, (void **)&keyMgr)) && keyMgr != nullptr) {
    keyMgr->AdviseKeyEventSink(clientId_, this, TRUE);
    keyMgr->Release();
  }
  ensureSession();
  return S_OK;
}

STDMETHODIMP TsimeTextService::Deactivate() {
  if (threadMgr_ != nullptr) {
    ITfKeystrokeMgr *keyMgr = nullptr;
    if (SUCCEEDED(threadMgr_->QueryInterface(IID_ITfKeystrokeMgr, (void **)&keyMgr)) && keyMgr != nullptr) {
      keyMgr->UnadviseKeyEventSink(clientId_);
      keyMgr->Release();
    }
    threadMgr_->Release();
    threadMgr_ = nullptr;
  }
  if (client_ != nullptr) {
    if (sessionId_[0] != '\0') {
      tsime_client_close_session(client_, sessionId_);
    }
    tsime_client_disconnect(client_);
    tsime_client_destroy(client_);
    client_ = nullptr;
    sessionId_[0] = '\0';
  }
  clientId_ = TF_CLIENTID_NULL;
  return S_OK;
}

bool TsimeTextService::ensureSession() {
  if (client_ != nullptr && sessionId_[0] != '\0') {
    return true;
  }
  client_ = tsime_client_create();
  if (client_ == nullptr) {
    return false;
  }
  const std::string path = pipePath();
  if (!tsime_client_connect(client_, path.c_str()) || !tsime_client_hello(client_, "windows-tsf")) {
    tsime_client_destroy(client_);
    client_ = nullptr;
    return false;
  }
  if (!tsime_client_open_session(client_, chineseMode_ ? TSIME_MODE_CHINESE : TSIME_MODE_ENGLISH, sessionId_,
                                 sizeof(sessionId_))) {
    tsime_client_disconnect(client_);
    tsime_client_destroy(client_);
    client_ = nullptr;
    return false;
  }
  return true;
}

std::string TsimeTextService::pipePath() {
  char buffer[256];
  DWORD n = GetEnvironmentVariableA("TSIME_PIPE", buffer, sizeof(buffer));
  if (n > 0 && n < sizeof(buffer)) {
    return std::string(buffer, n);
  }
  char user[128];
  DWORD userLen = sizeof(user);
  if (GetUserNameA(user, &userLen)) {
    return std::string("\\.\pipe\tsime-") + user;
  }
  return "\\.\pipe\tsime";
}
}

void TsimeTextService::applyResult(ITfContext *pic, const TsimeKeyResult &result) {
  if (pic == nullptr) {
    return;
  }
  if (result.commit_text != nullptr && result.commit_text[0] != '\0') {
    ITfInsertAtSelection *insert = nullptr;
    if (SUCCEEDED(pic->QueryInterface(IID_ITfInsertAtSelection, (void **)&insert)) && insert != nullptr) {
      ITfRange *range = nullptr;
      const std::wstring text = utf8ToWide(result.commit_text);
      insert->InsertTextAtSelection(0, TF_IAS_NOQUERY, text.c_str(), static_cast<LONG>(text.size()), &range);
      if (range != nullptr) {
        range->Release();
      }
      insert->Release();
    }
  }
}

STDMETHODIMP TsimeTextService::OnSetFocus(BOOL) { return S_OK; }

STDMETHODIMP TsimeTextService::OnTestKeyDown(ITfContext *, WPARAM wParam, LPARAM, BOOL *pfEaten) {
  if (pfEaten == nullptr) {
    return E_INVALIDARG;
  }
  *pfEaten = (!chineseMode_ || !isAsciiLetterOrEdit(wParam)) ? FALSE : TRUE;
  if (!chineseMode_) {
    *pfEaten = FALSE;
  }
  return S_OK;
}

STDMETHODIMP TsimeTextService::OnKeyDown(ITfContext *pic, WPARAM wParam, LPARAM, BOOL *pfEaten) {
  if (pfEaten == nullptr) {
    return E_INVALIDARG;
  }
  *pfEaten = FALSE;
  if (!chineseMode_) {
    return S_OK;
  }
  if (!ensureSession()) {
    return S_OK;
  }
  char chars[8] = {};
  uint32_t keyCode = 0;
  if (wParam >= 'A' && wParam <= 'Z') {
    chars[0] = static_cast<char>(wParam - 'A' + 'a');
  } else if (wParam >= '0' && wParam <= '9') {
    chars[0] = static_cast<char>(wParam);
  } else if (wParam == VK_SPACE) {
    chars[0] = ' ';
    keyCode = 49;
  } else if (wParam == VK_BACK) {
    keyCode = 51;
  } else if (wParam == VK_ESCAPE) {
    keyCode = 53;
  } else if (wParam == VK_DELETE) {
    keyCode = 117;
  }
  TsimeKeyResult result =
      tsime_client_handle_key(client_, sessionId_, keyCode, 0, chars, 1, 0, 0);
  if (result.ok && result.consumed) {
    *pfEaten = TRUE;
    applyResult(pic, result);
  }
  tsime_key_result_free(&result);
  return S_OK;
}

STDMETHODIMP TsimeTextService::OnTestKeyUp(ITfContext *, WPARAM, LPARAM, BOOL *pfEaten) {
  if (pfEaten) {
    *pfEaten = FALSE;
  }
  return S_OK;
}

STDMETHODIMP TsimeTextService::OnKeyUp(ITfContext *, WPARAM, LPARAM, BOOL *pfEaten) {
  if (pfEaten) {
    *pfEaten = FALSE;
  }
  return S_OK;
}

STDMETHODIMP TsimeTextService::OnPreservedKey(ITfContext *, REFGUID, BOOL *pfEaten) {
  if (pfEaten) {
    *pfEaten = FALSE;
  }
  return S_OK;
}

STDMETHODIMP TsimeClassFactory::QueryInterface(REFIID riid, void **ppvObj) {
  if (ppvObj == nullptr) {
    return E_INVALIDARG;
  }
  if (IsEqualIID(riid, IID_IUnknown) || IsEqualIID(riid, IID_IClassFactory)) {
    *ppvObj = static_cast<IClassFactory *>(this);
    AddRef();
    return S_OK;
  }
  *ppvObj = nullptr;
  return E_NOINTERFACE;
}

STDMETHODIMP_(ULONG) TsimeClassFactory::AddRef() { return InterlockedIncrement(&ref_); }

STDMETHODIMP_(ULONG) TsimeClassFactory::Release() {
  const LONG value = InterlockedDecrement(&ref_);
  if (value == 0) {
    delete this;
  }
  return value;
}

STDMETHODIMP TsimeClassFactory::CreateInstance(IUnknown *pUnkOuter, REFIID riid, void **ppvObject) {
  if (pUnkOuter != nullptr) {
    return CLASS_E_NOAGGREGATION;
  }
  auto *service = new TsimeTextService();
  const HRESULT hr = service->QueryInterface(riid, ppvObject);
  service->Release();
  return hr;
}

STDMETHODIMP TsimeClassFactory::LockServer(BOOL) { return S_OK; }
