{
  description = "Discord Show Local Time - A Discord RPC client to show local time";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, flake-utils }:
    flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = nixpkgs.legacyPackages.${system};
      in
      {
        packages.default = pkgs.buildGoModule {
          pname = "discord-show-local-time";
          version = "0.1.0";

          src = ./.;

          vendorHash = "sha256-tRwN1Jgdp+CVyRClnAbRQT54HC40hQ1Vk5BD11RzRcA=";

          # Build configuration
          subPackages = [ "." ];

          # Set the binary name to match what the Makefile expects
          postInstall = ''
            mv $out/bin/seelocaltime $out/bin/discord-time-presence
          '';

          meta = with pkgs.lib; {
            description = "A Discord RPC client to show local time in Discord status";
            homepage = "https://github.com/siddarthkay/discord-show-local-time";
            license = licenses.mit;
            maintainers = [ ];
            platforms = platforms.unix;
          };
        };

        # Development shell with Go and other tools
        devShells.default = pkgs.mkShell {
          buildInputs = with pkgs; [
            go
            gnumake
          ];

          shellHook = ''
            echo "Discord Show Local Time development environment"
            echo "Available commands:"
            echo "  make build    - Build the application"
            echo "  make run      - Run the application"
            echo "  make help     - Show all available make targets"
          '';
        };

        # Convenient apps
        apps.default = {
          type = "app";
          program = "${self.packages.${system}.default}/bin/discord-time-presence";
        };
      });
}