#include "tsime_engine.h"

#include <fcitx/inputpanel.h>
#include <fcitx/candidatelist.h>

#include <cstdlib>
#include <unistd.h>

namespace fcitx {
namespace {

class TsimeCandidateWord : public CandidateWord {
public:
  TsimeCandidateWord(TsimeEngine *engine, std::string text, int index)
      : CandidateWord(Text(std::move(text))), engine_(engine), index_(index) {}

  void select(InputContext *inputContext) const override {
    if (engine_ == nullptr || inputContext == nullptr) {
      return;
    }
    engine_->selectIndex(inputContext, index_);
  }

private:
  TsimeEngine *engine_ = nullptr;
  int index_ = 0;
};

} // namespace

void TsimeEngine::selectIndex(InputContext *ic, int index) {
  if (!ensureSession() || client_ == nullptr) {
    return;
  }
  TsimeKeyResult result = tsime_client_select_candidate(client_, sessionId_, 0, static_cast<uint32_t>(index));
  applyResult(ic, result);
  tsime_key_result_free(&result);
}

TsimeEngine::TsimeEngine(Instance *instance) : instance_(instance) {}

TsimeEngine::~TsimeEngine() {
  if (client_ != nullptr) {
    if (sessionId_[0] != '\0') {
      tsime_client_close_session(client_, sessionId_);
    }
    tsime_client_disconnect(client_);
    tsime_client_destroy(client_);
    client_ = nullptr;
  }
}

std::string TsimeEngine::socketPath() {
  if (const char *overridePath = std::getenv("TSIME_SOCKET")) {
    if (overridePath[0] != '\0') {
      return overridePath;
    }
  }
  if (const char *runtime = std::getenv("XDG_RUNTIME_DIR")) {
    if (runtime[0] != '\0') {
      return std::string(runtime) + "/tsime.sock";
    }
  }
  return std::string("/tmp/tsime-") + std::to_string(getuid()) + ".sock";
}

bool TsimeEngine::ensureSession() {
  if (client_ != nullptr && sessionId_[0] != '\0') {
    return true;
  }
  client_ = tsime_client_create();
  if (client_ == nullptr) {
    return false;
  }
  const auto path = socketPath();
  if (!tsime_client_connect(client_, path.c_str()) || !tsime_client_hello(client_, "linux-fcitx5")) {
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

void TsimeEngine::applyResult(InputContext *ic, const TsimeKeyResult &result) {
  if (ic == nullptr) {
    return;
  }
  if (result.commit_text != nullptr && result.commit_text[0] != '\0') {
    ic->commitString(result.commit_text);
  }
  Text preedit;
  if (result.preedit != nullptr) {
    preedit.append(result.preedit, TextFormat::Underline);
  }
  ic->inputPanel().reset();
  ic->inputPanel().setClientPreedit(preedit);
  if (result.candidate_count > 0 && result.candidates != nullptr) {
    auto candidateList = std::make_unique<CommonCandidateList>();
    candidateList->setSelectionKey(KeyList{
        Key("1"), Key("2"), Key("3"), Key("4"), Key("5"), Key("6"), Key("7"), Key("8"), Key("9"),
    });
    for (size_t i = 0; i < result.candidate_count; ++i) {
      const char *text = result.candidates[i].text ? result.candidates[i].text : "";
      candidateList->append(std::make_unique<TsimeCandidateWord>(this, text, static_cast<int>(i)));
    }
    ic->inputPanel().setCandidateList(std::move(candidateList));
  }
  ic->updatePreedit();
  ic->updateUserInterface(UserInterfaceComponent::InputPanel);
}

void TsimeEngine::activate(const InputMethodEntry &, InputContextEvent &) { ensureSession(); }

void TsimeEngine::deactivate(const InputMethodEntry &, InputContextEvent &event) {
  if (auto *ic = event.inputContext()) {
    ic->inputPanel().reset();
    ic->updatePreedit();
  }
}

void TsimeEngine::reset(const InputMethodEntry &, InputContextEvent &event) {
  if (auto *ic = event.inputContext()) {
    ic->inputPanel().reset();
    ic->updatePreedit();
  }
}

void TsimeEngine::keyEvent(const InputMethodEntry &, KeyEvent &keyEvent) {
  if (keyEvent.isRelease()) {
    return;
  }
  if (!chineseMode_) {
    return;
  }
  auto *ic = keyEvent.inputContext();
  if (ic == nullptr || !ensureSession()) {
    return;
  }
  const auto &key = keyEvent.key();
  std::string characters;
  uint32_t keyCode = 0;
  if (key.isSimple()) {
    characters = keyEvent.key().toString();
  }
  if (key.check(FcitxKey_BackSpace)) {
    keyCode = 51;
    characters.clear();
  } else if (key.check(FcitxKey_Delete)) {
    keyCode = 117;
    characters.clear();
  } else if (key.check(FcitxKey_Escape)) {
    keyCode = 53;
    characters.clear();
  } else if (key.check(FcitxKey_space)) {
    keyCode = 49;
    characters = " ";
  }
  TsimeKeyResult result =
      tsime_client_handle_key(client_, sessionId_, keyCode, 0, characters.c_str(), 1, 0, 0);
  if (result.ok && result.consumed) {
    keyEvent.filterAndAccept();
    applyResult(ic, result);
  }
  tsime_key_result_free(&result);
}

AddonInstance *TsimeEngineFactory::create(AddonManager *manager) {
  return new TsimeEngine(manager->instance());
}

} // namespace fcitx

FCITX_ADDON_FACTORY(fcitx::TsimeEngineFactory);
