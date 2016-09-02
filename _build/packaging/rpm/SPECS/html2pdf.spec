Name:           %{_product_name}
Version:        %{_product_version}

%if 0%{?rhel} >= 5
Release:        1.el%{?rhel}
%else
Release:        1%{?dist}
%endif

Summary:        Generating pdf from html
Group:          Development/Tools
License:        MIT
Source0:        %{name}_linux_amd64.zip
BuildRoot:      %(mktemp -ud %{_tmppath}/%{name}-%{version}-%{release}-XXXXXX)

%description
Generating pdf from html

%prep
%setup -q -c

%install
mkdir -p %{buildroot}/%{_bindir}
cp %{name} %{buildroot}/%{_bindir}

%pre

%post

%preun

%clean
rm -rf %{buildroot}

%files
%defattr(-,root,root,-)
%attr(755, root, root) %{_bindir}/%{name}

%doc
