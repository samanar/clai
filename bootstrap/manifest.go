package bootstrap

type Asset struct {
	URL      string
	Filename string
}

type Manifest struct {
	Llama Asset
	Model Asset
}

func DefaultManifest() Manifest {
	var llama Asset

	llama = Asset{
		URL:      "https://github.com/Mozilla-Ocho/llamafile/releases/download/0.9.3/llamafile-0.9.3",
		Filename: "llamafile",
	}

	model := Asset{
		URL:      "https://huggingface.co/Mozilla/gemma-3-1b-it-llamafile/resolve/main/google_gemma-3-1b-it-Q6_K.llamafile?download=true",
		Filename: "gemma-3-1b-it-q6.llamafile",
	}
	return Manifest{Llama: llama, Model: model}
}
