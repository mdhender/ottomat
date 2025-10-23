package templates

import "fmt"

func Footer(version string) string {
	return fmt.Sprintf(`
    <footer class="bg-gray-800 text-gray-400 text-sm py-4 mt-8 border-t border-gray-700">
        <div class="container mx-auto px-8 flex justify-between items-center">
            <div>
                Version %s
            </div>
            <div class="space-x-4">
                <a href="https://github.com/mdhender/ottomat" target="_blank" class="hover:text-white transition">GitHub</a>
                <a href="https://discord.gg/8v2pWUs2Pg" target="_blank" class="hover:text-white transition">Discord</a>
            </div>
        </div>
    </footer>`, version)
}
