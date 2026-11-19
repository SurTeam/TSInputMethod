#pragma once

#include <fcitx/addonfactory.h>
#include <fcitx/addonmanager.h>
#include <fcitx/inputmethodengine.h>
#include <fcitx/instance.h>

#include "tsime_c_api.h"

namespace fcitx {

class TsimeEngine : public InputMethodEngineV2 {
public:
  explicit TsimeEngine(Instance *instance);
  ~TsimeEngine() override;

  void keyEvent(const InputMethodEntry &entry, KeyEvent &keyEvent) override;
  void reset(const InputMethodEntry &entry, InputContextEvent &event) override;
  void activate(const InputMethodEntry &entry, InputContextEvent &event) override;
  void deactivate(const InputMethodEntry &entry, InputContextEvent &event) override;

  void selectIndex(InputContext *ic, int index);

private:
  bool ensureSession();
  void applyResult(InputContext *ic, const TsimeKeyResult &result);
  static std::string socketPath();

  Instance *instance_ = nullptr;
  TsimeClient *client_ = nullptr;
  char sessionId_[128] = {};
  bool chineseMode_ = true;
};

class TsimeEngineFactory : public AddonFactory {
public:
  AddonInstance *create(AddonManager *manager) override;
};

} // namespace fcitx
