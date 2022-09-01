{
  description = "Vendorito";

  inputs = {
    nixpkgs.url = "nixpkgs";
    utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, utils }:
    utils.lib.eachDefaultSystem (system:
      let
        trim = s: nixpkgs.lib.strings.removePrefix " v" (nixpkgs.lib.strings.removeSuffix " " (nixpkgs.lib.removeSuffix "\n" s));
        lastModifiedDate = self.lastModifiedDate or self.lastModified or "19700101";
        version = trim (builtins.readFile ./VERSION);
        pkgs = nixpkgs.legacyPackages.${system};
      in
      {
        packages.default = pkgs.buildGoModule {
          pname = "vendorito";
          version = "v${version}";
          src = ./.;
          subPackages = [ "cmd/vendorito" ];
          buildInputs = with pkgs; [ gpgme.dev ];

          # vendorSha256 = pkgs.lib.fakeSha256;
          vendorSha256 = "sha256-SGfB2v/vVlOFv+hJrzR60UTTf7vgAkokoi2HCDEew4I=";
        };

        apps.default = utils.lib.mkApp { drv = self.packages.${system}.default; };
        devShells.default = pkgs.mkShell
          {
            buildInputs = with pkgs; [
              go_1_18
              gotools
              goreleaser
              just
              gpgme.dev
            ];
          };
      });
}
