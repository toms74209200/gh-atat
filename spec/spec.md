TODO.md の内容を GitHub の Issues に反映するCLIアプリケーション

git と同様の操作体系

## TODO.md の内容を GitHub の Issues に push する

```bash
gh atat push
```

- TODO.md にある未チェックの項目が GitHub の Issues に登録されていないとき, Issue を新規作成する
- GitHub の Issues にある open な Issue のうち, TODO.md にある項目がチェックされているものは, Issue をクローズする

1. TODO.mdの項目がGitHub Issuesにない場合

 - 新規Issueを作成
 - 作成したIssue番号をTODO.mdに追記

2. TODO.mdの項目がGitHub Issuesに既にある場合（タイトルが一致）

 - 既存のIssue番号をTODO.mdに追記

## GitHub の Issues から TODO.md の内容を更新する

```bash
gh atat pull
```

- GitHub の Issues にある open な Issue が TODO.md にないとき, TODO.md に追加する
- TODO.md にある未チェックの項目が GitHub の Issues ではクローズされているとき, TODO.md の項目をチェックする

## Issue内容の同期範囲

以下の情報のみを同期対象とする:
- タイトル: TODO.mdの項目テキストとIssueのタイトルを同期
- 状態: TODO.mdのチェック状態とIssueのopen/closed状態を同期
- Issue番号: TODO.mdの項目に対応するIssue番号を記録

## TODO.mdの構造

- 階層構造（ネスト）は扱わない。すべての項目をフラットな構造として扱う
- チェックボックス形式の項目のみを同期対象とする

## 実装

GitHub CLI (gh) の拡張機能として実装する

https://docs.github.com/en/github-cli/github-cli/creating-github-cli-extensions

- GitHub APIの認証方法
  - GitHub CLI (`gh`) の認証を使用
    - ユーザーは事前に `gh auth login` で認証
    - gh-atatは `gh api` コマンドを使用してGitHub APIにアクセス
    - 認証トークンの管理はGitHub CLIが行う
- 必要な権限スコープ
  - `repo`: リポジトリへのフルアクセス（Issuesの作成、更新、クローズを含む）
- コマンド実行例
  ```bash
  # インストール
  gh extension install owner/gh-atat

  # 実行（gh auth loginが必要）
  gh atat push
  gh atat pull
  gh atat remote
  ```
- コマンド出力例
  ```
  $ gh atat push
  X Authentication required
  ℹ To get started with GitHub CLI, please run:  gh auth login
  ```

## リポジトリ設定
- リポジトリ設定は git remote のように以下のサブコマンドで管理する:
  ```bash
  # リポジトリの追加
  $ gh atat remote add owner/repo
  ✓ Repository owner/repo has been added

  # 現在の設定を表示
  $ gh atat remote
  owner/repo

  # リポジトリの削除
  $ gh atat remote remove owner/repo
  ✓ Repository owner/repo has been removed
  ```
- 設定は ~/.config/gh-atat/config.json に保存
- 複数プロジェクトの場合は、.git/config のように、.gh-atat/config でプロジェクト固有の設定を上書き可能
