#pragma once

#include <windows.h>
#include <msctf.h>

#include <string>

#include "tsime_c_api.h"

class TsimeTextService : public ITfTextInputProcessorEx, public ITfKeyEventSink {
public:
  TsimeTextService();
  virtual ~TsimeTextService();

  STDMETHODIMP QueryInterface(REFIID riid, void **ppvObj) override;
  STDMETHODIMP_(ULONG) AddRef() override;
  STDMETHODIMP_(ULONG) Release() override;

  STDMETHODIMP Activate(ITfThreadMgr *ptim, TfClientId tid) override;
  STDMETHODIMP ActivateEx(ITfThreadMgr *ptim, TfClientId tid, DWORD dwFlags) override;
  STDMETHODIMP Deactivate() override;

  STDMETHODIMP OnSetFocus(BOOL fForeground) override;
  STDMETHODIMP OnTestKeyDown(ITfContext *pic, WPARAM wParam, LPARAM lParam, BOOL *pfEaten) override;
  STDMETHODIMP OnKeyDown(ITfContext *pic, WPARAM wParam, LPARAM lParam, BOOL *pfEaten) override;
  STDMETHODIMP OnTestKeyUp(ITfContext *pic, WPARAM wParam, LPARAM lParam, BOOL *pfEaten) override;
  STDMETHODIMP OnKeyUp(ITfContext *pic, WPARAM wParam, LPARAM lParam, BOOL *pfEaten) override;
  STDMETHODIMP OnPreservedKey(ITfContext *pic, REFGUID rguid, BOOL *pfEaten) override;

private:
  bool ensureSession();
  void applyResult(ITfContext *pic, const TsimeKeyResult &result);
  static std::string pipePath();

  LONG ref_ = 1;
  ITfThreadMgr *threadMgr_ = nullptr;
  TfClientId clientId_ = TF_CLIENTID_NULL;
  TsimeClient *client_ = nullptr;
  char sessionId_[128] = {};
  bool chineseMode_ = true;
};

class TsimeClassFactory : public IClassFactory {
public:
  STDMETHODIMP QueryInterface(REFIID riid, void **ppvObj) override;
  STDMETHODIMP_(ULONG) AddRef() override;
  STDMETHODIMP_(ULONG) Release() override;
  STDMETHODIMP CreateInstance(IUnknown *pUnkOuter, REFIID riid, void **ppvObject) override;
  STDMETHODIMP LockServer(BOOL fLock) override;

private:
  LONG ref_ = 1;
};

extern const CLSID CLSID_TsimeTextService;
