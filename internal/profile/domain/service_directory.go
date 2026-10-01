package domain

import "slices"

// Служебные каталоги — закрытый перечень имён, которые пустоте целевого каталога не мешают.
// Перечень закрыт намеренно: правило «игнорировать всё скрытое» пропустило бы под перезапись
// и `.gitignore`, и манифест.
const (
	serviceDirGit     = ".git"
	serviceDirMercury = ".hg"
	serviceDirSVN     = ".svn"
	serviceDirIDEA    = ".idea"
	serviceDirVSCode  = ".vscode"
	serviceDirDSStore = ".DS_Store"
)

// ServiceDirectories возвращает закрытый перечень служебных каталогов. Решение «что считать
// помехой пустоте» принадлежит домену: оно не зависит ни от порядка шагов опроса, ни от файловой
// системы, поэтому проверяется чистым тестом.
func ServiceDirectories() []string {
	return []string{
		serviceDirGit,
		serviceDirMercury,
		serviceDirSVN,
		serviceDirIDEA,
		serviceDirVSCode,
		serviceDirDSStore,
	}
}

// BlocksEmptiness отвечает, мешает ли элемент с таким именем считать каталог пустым. Элемент вне
// перечня служебных каталогов мешает независимо от того, скрытый он или нет.
func BlocksEmptiness(name string) bool {
	return !slices.Contains(ServiceDirectories(), name)
}
